package tactics

import (
	"fmt"
)

// "log"
// "net/http"
// "github.com/gorilla/mux"

// func YourHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("adsf"))
// }

func main() {
	// myString := "thisIsAStringOfWordsInCamelCaseButICanAddAFewMoreWords"
	var input string
	wordCount := 1

	fmt.Scanf("%s\n", &input)

	for _, ch := range input {
		if ch > 64 && ch < 91 {
			wordCount++
		}
	}

	fmt.Printf("There are %d words in the camel-case string, \"%s\".\n", wordCount, input)

	// myImage := image.NewRGBA(image.Rect(0, 0, 100, 200))
	//
	// outputFile, err := os.Create("test.png")
	// if err != nil {
	// 	fmt.Println("ERROR!!!")
	// }
	// defer outputFile.Close()
	//
	// png.Encode(outputFile, myImage)

	// r := mux.NewRouter()
	// r.HandleFunc("/", YourHandler)
	// log.Fatal(http.ListenAndServe(":8000", r))
}
