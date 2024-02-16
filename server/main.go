package main

// ===== Imports =====
import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"sort"
	"unicode/utf8"
)
// ==========



// ===== Variables globales =====
const BASE_ATTEMPTS = 10
const BASE_PORT = "8080"

var HangManEasy HangManData
var HangMan HangManData
var HangManHard HangManData
// ==========



// ===== STRUCT =====

// Type HangManData
type HangManData struct {
	Word             string
	ToFind           string
	Attempts         int
	HangmanPositions []string
	Points 			 int
	WordUsed		 []string
}

// Type Page
type Page struct {
	HangmanDraw string
	Attempts    int
	Word        string
	Leaderboard	template.HTML
	Points		int
	WordUsed	[]string
}

// Type UserLeaderBoard pour le classement via JSON
type UserLeaderBoard struct {
	Username string `json:"username"`
	Score    int    `json:"score"`
}
//==========



// ===== FONCTION =====

// Fonction Remove CR (Carriage Return)
func removeCR(str string) string {
	r:=""
	for _,c := range str{
		if c!=13{
			r+=string(c)
		}
	}
	return r
}

// Fonction traitement du jeu HangMan pour différents niveaux
func hangManGame(HangMan *HangManData, playerChoice string, mode int, r *http.Request) (*Page, *template.Template) {

	if r.Method == http.MethodPost{
		playerChoiceLen:=utf8.RuneCountInString(playerChoice)

		//Si le mot/lettre a déjà été tenté
		isUsed := false
		for i,c := range HangMan.WordUsed{
			if c==playerChoice{
				isUsed=true
			} else if i==len(HangMan.WordUsed)-1 && c!=playerChoice{
				HangMan.WordUsed=append(HangMan.WordUsed, playerChoice)
			}
		}

		if !isUsed{
			if playerChoiceLen==1{
				success := false
				for i, c := range HangMan.ToFind {
					if string(c) == playerChoice {
						HangMan.Word = HangMan.Word[:i] + string(c) + HangMan.Word[i+1:]
						success = true
					}
				}
				if playerChoiceLen != 0 && !success && HangMan.Attempts-1>=0{
					if mode==2{
						HangMan.Attempts-=2
					} else{
						HangMan.Attempts--
					}
				}
			} else if playerChoiceLen!=0 {
				if playerChoice!=HangMan.ToFind{
					if HangMan.Attempts-2<0{
						HangMan.Attempts=0
					} else {
						if mode==0{
							HangMan.Attempts--
						} else {
							HangMan.Attempts-=2
						}
					}
				} else {
					HangMan.Word=HangMan.ToFind
				}
			}
		}
	}
		
	//=== Config de la page ===
	p := &Page{}
	if HangMan.Attempts == BASE_ATTEMPTS {
		p = &Page{
			HangmanDraw: "",
			Attempts:    HangMan.Attempts,
			Word:        displayWordHide(HangMan.Word),
			WordUsed: 	 HangMan.WordUsed,
		}
	} else {
		p = &Page{
			HangmanDraw: HangMan.HangmanPositions[len(HangMan.HangmanPositions)-1-HangMan.Attempts],
			Attempts:    HangMan.Attempts,
			Word:        displayWordHide(HangMan.Word),
			WordUsed: 	 HangMan.WordUsed,
		}
	}
	//======

	//=== Page à afficher ===
	var t *template.Template
	
	if mode==0{
		t,_ = template.ParseFiles("../public/hangmanEasy.html")
	} else if mode==1{
		t,_ = template.ParseFiles("../public/hangmanNormal.html")
	} else{
		t,_ = template.ParseFiles("../public/hangmanHard.html")
	}

	if HangMan.Word == HangMan.ToFind {
		if mode==0{
			HangMan.Points=(len(HangMan.ToFind)-((BASE_ATTEMPTS-HangMan.Attempts)/BASE_ATTEMPTS*len(HangMan.ToFind)))*80
		} else if mode==1{
			HangMan.Points=(len(HangMan.ToFind)-((BASE_ATTEMPTS-HangMan.Attempts)/BASE_ATTEMPTS*len(HangMan.ToFind)))*100
		} else {
			HangMan.Points=(len(HangMan.ToFind)-((BASE_ATTEMPTS-HangMan.Attempts)/BASE_ATTEMPTS*len(HangMan.ToFind)))*150
		}
		
		p.Points=HangMan.Points
		t, _ = template.ParseFiles("../public/winning.html")
	}else if HangMan.Attempts == 0 {
		p.Word=HangMan.ToFind
		t, _ = template.ParseFiles("../public/losing.html")
	}
	//======

	return p,t
}

// Fonction de lecture de fichier JSON
func writeJSONToFile(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
// Fonction d'écriture de fichier JSON
func readJSONFromFile(filename string) ([]UserLeaderBoard, error) {
	var users []UserLeaderBoard

	file, err := os.Open(filename)
	if err != nil {
		return users, err
	}
	defer file.Close()

	jsonData, err := ioutil.ReadAll(file)
	if err != nil {
		return users, err
	}

	err = json.Unmarshal(jsonData, &users)
	if err != nil {
		return users, err
	}

	return users, nil
}

// Fonction de traitement du mot caché pour avoir un joli texte avec espacement
func displayWordHide(word string) string {
	r:=""
	for i:=0; i<len(word)-1; i++{
		r+=string(word[i])+" "
	}
	r+=string(word[len(word)-1])
	return r
}

// Fonction de réinitialisation du HangMan en fonction des différents niveaux
func resetHangman(HangMan *HangManData, mode int) {
	//=== Assignation des 10 essais et des essais ===
	HangMan.Attempts = BASE_ATTEMPTS
	HangMan.WordUsed = []string{}
	//======

	//=== Ouverture du fichier hangman.txt et ajout des différentes positions à notre type HangManData ===
	content, err := os.ReadFile("hangman.txt")
	if err != nil {
		log.Fatal(err)
	}
	HangMan.HangmanPositions = []string(strings.Split(string(content), ","))
	//======

	//=== Ouverture du fichier words.txt et choix du mot en fonction du niveau ===
	if mode==0{
		content, err = os.ReadFile("words1.txt")
		if err != nil {
			log.Fatal(err)
		}
	} else if mode == 1 {
		content, err = os.ReadFile("words2.txt")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		content, err = os.ReadFile("words3.txt")
		if err != nil {
			log.Fatal(err)
		}
	}

	words := strings.Split(string(content), "\n")
	HangMan.ToFind = removeCR(strings.ToUpper(words[rand.Intn(len(words))]))
	//======

	//=== Création du mot vide et ajout des lettres en fonction du niveau de difficulté ===
	wordHidden := ""
	if mode==0{
		midLetter := string(HangMan.ToFind[len(HangMan.ToFind)/2])
		lastLetter := string(HangMan.ToFind[len(HangMan.ToFind)-1])

		HangMan.WordUsed=append(HangMan.WordUsed, midLetter)
		if midLetter!=lastLetter{
			HangMan.WordUsed=append(HangMan.WordUsed, lastLetter)
		}
		
		for i := 0; i < len(HangMan.ToFind)-1; i++ {
			if string(HangMan.ToFind[i]) == lastLetter{
				wordHidden += lastLetter
			} else if string(HangMan.ToFind[i])==midLetter{
				wordHidden += midLetter
			} else {
				wordHidden += "_"
			}
		}
		wordHidden += lastLetter
	} else if mode==1{
		lastLetter := string(HangMan.ToFind[len(HangMan.ToFind)-1])
		HangMan.WordUsed=append(HangMan.WordUsed, lastLetter)

		for i := 0; i < len(HangMan.ToFind)-1; i++ {
			if string(HangMan.ToFind[i]) == lastLetter {
				wordHidden += lastLetter
			} else {
				wordHidden += "_"
			}
		}
		wordHidden += lastLetter
	} else{
		for i := 0; i < len(HangMan.ToFind); i++ {
			wordHidden += "_"
		}
	}
	HangMan.Word = wordHidden
	//======

	//=== Points par défaut ===
	HangMan.Points=0
	//======
}
//==========



//===== HANDLERS =====

// Home handler
func HomePage(w http.ResponseWriter, r *http.Request) {
	p := &Page{}
	t, _ := template.ParseFiles("../public/index.html")

	t.Execute(w, p)
}

// Ajout de score au classement
func AddScorePage(w http.ResponseWriter, r *http.Request) {
	playerName := r.FormValue("input")
	if len(playerName)>0 && (HangMan.Word==HangMan.ToFind || HangManHard.Word==HangManHard.ToFind || HangManEasy.Word==HangManEasy.ToFind){
		usersLeaderBoard, err := readJSONFromFile("leaderboard.json")
		if err != nil {
			fmt.Println("Erreur lors de la lecture du fichier JSON:", err)
			return
		}

		points:=0
		if HangMan.Word==HangMan.ToFind{
			points=HangMan.Points
		} else if HangManHard.Word==HangManHard.ToFind{
			points=HangManHard.Points
		} else if HangManEasy.Word==HangManEasy.ToFind{
			points=HangManEasy.Points
		}
	
		usersLeaderBoard=append(usersLeaderBoard, UserLeaderBoard{
			Username: playerName,
			Score:    points,
		})
	
		err = writeJSONToFile("leaderboard.json", usersLeaderBoard)
		if err != nil {
			fmt.Println("Erreur lors de l'écriture du fichier JSON:", err)
			return
		}
	}

	p := &Page{}
	t, _ := template.ParseFiles("../public/index.html")

	t.Execute(w, p)
}

// Leaderboard/Classement handler
func LeaderboardPage(w http.ResponseWriter, r *http.Request) {
	//Lecture des utilisateurs dans le fichier leaderboard.json
	readUsers, err := readJSONFromFile("leaderboard.json")
	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier JSON:", err)
		return
	}

	//Tri dans l'ordre décroissant des utilisateurs en fonction de leur score
	sort.Slice(readUsers, func(i, j int) bool {
		return readUsers[i].Score > readUsers[j].Score
	})

	//Affichage des utilisateurs en HTML en fonction de leur classement
	leaderboard:=""
	for i, user := range readUsers {
		if i==5{
			break
		}
		if i==0{
			leaderboard+="<div class='box' style='background-color:#FFD433'>"+user.Username+" : "+strconv.Itoa(user.Score)+"</div>"
		} else if i==1{
			leaderboard+="<div class='box' style='background-color:#D7D7D7'>"+user.Username+" : "+strconv.Itoa(user.Score)+"</div>"
		} else if i==2{
			leaderboard+="<div class='box' style='background-color:#F59854'>"+user.Username+" : "+strconv.Itoa(user.Score)+"</div>"
		} else {
			leaderboard+="<div class='box'>"+user.Username+" : "+strconv.Itoa(user.Score)+"</div>"
		}
		
	}

	//Configuration de la page
	p := &Page{
		Leaderboard:  template.HTML(leaderboard),
	}
	t, _ := template.ParseFiles("../public/gameleaderboard.html")

	t.Execute(w, p)
}

// Hangman EASY handler
func HangmanEasyPage(w http.ResponseWriter, r *http.Request) {
	playerChoice := strings.ToUpper(r.FormValue("input"))
	p,t := hangManGame(&HangManEasy, playerChoice, 0, r)
	t.Execute(w, p)
}

// Hangman NORMAL handler
func HangmanNormalPage(w http.ResponseWriter, r *http.Request) {
	playerChoice := strings.ToUpper(r.FormValue("input"))
	p,t := hangManGame(&HangMan, playerChoice, 1, r)
	t.Execute(w, p)
}

// Hangman HARD handler
func HangmanHardPage(w http.ResponseWriter, r *http.Request) {
	playerChoice := strings.ToUpper(r.FormValue("input"))
	p,t := hangManGame(&HangManHard, playerChoice, 2, r)
	t.Execute(w, p)
}


// Reset EASY handler
func ResetEasyHandler(w http.ResponseWriter, r *http.Request) {
	resetHangman(&HangManEasy, 0)

	p,t := hangManGame(&HangManEasy, "", 0, r)
	t.Execute(w, p)
}

// Reset NORMAL handler
func ResetHandler(w http.ResponseWriter, r *http.Request) {
	resetHangman(&HangMan, 1)

	p,t := hangManGame(&HangMan, "", 1, r)
	t.Execute(w, p)
}

// Reset HARD handler
func ResetHardHandler(w http.ResponseWriter, r *http.Request) {
	resetHangman(&HangManHard, 2)

	p,t := hangManGame(&HangManHard, "", 2, r)
	t.Execute(w, p)
}
//==========



func main() {
	//=== Lancement du jeu ===
	resetHangman(&HangManEasy, 0)
	resetHangman(&HangMan, 1)
	resetHangman(&HangManHard, 2)
	//======

	//=== Accès aux fichiers statiques pour le client ===
	fs := http.FileServer(http.Dir("../static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	//======

	//=== Handlers ===
	http.HandleFunc("/", HomePage)

	http.HandleFunc("/hangmanEasy", HangmanEasyPage)
	http.HandleFunc("/resetEasy", ResetEasyHandler)

	http.HandleFunc("/hangman", HangmanNormalPage)
	http.HandleFunc("/reset", ResetHandler)

	http.HandleFunc("/hangmanHard", HangmanHardPage)
	http.HandleFunc("/resetHard", ResetHardHandler)

	http.HandleFunc("/classement", LeaderboardPage)
	http.HandleFunc("/addScore", AddScorePage)
	//======

	//=== Lancement du serveur web ===
	fmt.Println("Serveur web lancé sur http://127.0.0.1:" + BASE_PORT)
	log.Fatal(http.ListenAndServe(":"+BASE_PORT, nil))
	//======
}
