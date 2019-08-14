package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

func main() {

	data := flag.String("f", "default", "usage")
	//data2 := flag.String("s", "default", "usage")
	flag.Parse()
	// can't upload when open in browser tab cause file is open
	genaudio(data, "output.mp3")
	//genaudio(data2, "output2.mp3")
	upload("http://127.0.0.1:3030/upload", "output.mp3")
	//upload("http://127.0.0.1:3030/upload", "output2.mp3")
}

func genaudio(words *string, filename string) {
	// Instantiates a client.
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Perform the text-to-speech request on the text input with the selected
	// voice parameters and audio file type.
	req := texttospeechpb.SynthesizeSpeechRequest{
		// Set the text input to be synthesized.
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: (*words)},
		},
		// Build the voice request, select the language code ("en-US") and the SSML
		// voice gender ("neutral").
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-US",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
		},
		// Select the type of audio file you want returned.
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Fatal(err)
	}

	// The resp's AudioContent is binary.
	filenameinput := filename
	err = ioutil.WriteFile(filenameinput, resp.AudioContent, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Audio content written to file: %v\n", filenameinput)
}

func upload(upurl string, filename string) {
	file, err := os.Open("./" + filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	res, err := http.Post(upurl, "binary/octet-stream", file)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	message, _ := ioutil.ReadAll(res.Body)
	fmt.Printf(string(message))
}
