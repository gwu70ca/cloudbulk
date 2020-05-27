package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var gcpHeader = `First Name [Required],Last Name [Required],Email Address [Required],Password [Required],Password Hash Function [UPLOAD ONLY],Org Unit Path [Required],New Primary Email [UPLOAD ONLY],Status [READ ONLY],Last Sign In [READ ONLY],Recovery Email,Home Secondary Email,Work Secondary Email,Recovery Phone [MUST BE IN THE E.164 FORMAT],Work Phone,Home Phone,Mobile Phone,Work Address,Home Address,Employee ID,Employee Type,Employee Title,Manager Email,Department,Cost Center,2sv Enrolled [READ ONLY],2sv Enforced [READ ONLY],Building ID,Floor Name,Floor Section,Email Usage [READ ONLY],Drive Usage [READ ONLY],Total Storage [READ ONLY],Change Password at Next Sign-In,New Status [UPLOAD ONLY]`

func check(ptr *string, msg string) bool {
	//fmt.Println(*ptr)
	if *ptr == "" {
		fmt.Println(msg)
		return false
	}
	return true
}

func main() {
	var source, target, sourceFile, targetFile string
	flag.StringVar(&source, "source", "gcp", "Source IdP")
	flag.StringVar(&target, "target", "azure", "Target IdP")
	flag.StringVar(&sourceFile, "sourceFile", "", "CSV of source IdP")
	flag.StringVar(&targetFile, "targetFile", "", "CSV of target IdP")

	flag.Parse()

	if !check(&source, "Need source IdP name: [AWS, Azure, GCP") {
		return
	}

	if !check(&target, "Need target IdP name: [AWS, Azure, GCP") {
		return
	}

	if !check(&sourceFile, "Need source CSV file name") {
		return
	}

	if !check(&targetFile, "Need target CSV file name") {
		return
	}

	fmt.Println("Source IdP: ", source)
	fmt.Println("Target IdP: ", target)
	fmt.Println("Source CSV File: ", sourceFile)
	fmt.Println("Target CSV File: ", targetFile)

	if source == "gcp" {
		readGcpUsr(sourceFile)
	}
}

func readGcpUsr(gcpFile string) {
	file, err := os.Open(gcpFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	h := strings.Split(gcpHeader, ",")

	r := csv.NewReader(bufio.NewReader(file))

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		for idx, f := range record {
			fmt.Println(h[idx], "=", f)
		}

	}
}
