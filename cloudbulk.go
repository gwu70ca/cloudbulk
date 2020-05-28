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
var azureHeader = `userPrincipalName,displayName,surname,mail,givenName,objectId,userType,jobTitle,department,accountEnabled,usageLocation,streetAddress,state,country,physicalDeliveryOfficeName,city,postalCode,telephoneNumber,mobile,authenticationPhoneNumber,authenticationAlternativePhoneNumber,authenticationEmail,alternateEmailAddress,ageGroup,consentProvidedForMinor,legalAgeGroupClassification`

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

	var sourceRecords []map[string]string
	if source == "gcp" {
		sourceRecords = readGcpUsr(sourceFile)
		if sourceRecords == nil || len(sourceRecords) == 0 {
			fmt.Println("Can not read source CSV file or it is empty")
			return
		}
	} else if source == "azure" {

	}

	if target == "azure" {
		mappingFile := source + "_" + target + ".txt"
		writeAzureUser(sourceRecords, mappingFile)
	} else if target == "aws" {

	}
}

func writeAzureUser(sourceRecords []map[string]string, mappingFile string) {
	f, err := os.Open(mappingFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	//Create header mapping between two formats
	headerMap := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		ss := strings.Split(line, "=")
		if len(ss) > 1 && len(ss[0]) > 0 && len(ss[1]) > 0 {
			headerMap[ss[1]] = ss[0]
		}
	}

	for k, v := range headerMap {
		fmt.Println("mapping", k, "to", v)
	}

	var newRecords []map[string]string
	azureHeaders := strings.Split(azureHeader, ",")
	for _, sourceRerord := range sourceRecords {
		newRecord := make(map[string]string)

		for _, header := range azureHeaders {
			sourceHeader := headerMap[header]
			if sourceHeader != "" {
				value := sourceRerord[sourceHeader]
				if value != "" {
					newRecord[header] = value
				}
			}
		}

		newRecords = append(newRecords, newRecord)
	}

	fmt.Println("result: ")
	fmt.Println("==================================================")
	for _, newRecord := range newRecords {
		fmt.Println(newRecord)
	}
}

func readGcpUsr(gcpFile string) []map[string]string {
	file, err := os.Open(gcpFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()

	h := strings.Split(gcpHeader, ",")

	r := csv.NewReader(bufio.NewReader(file))

	lineNumber := 0
	var records []map[string]string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		lineNumber = lineNumber + 1
		//Skip header line
		if lineNumber == 0 {
			continue
		}

		m := make(map[string]string)
		for idx, r := range record {
			if r == "" {
				continue
			}
			fmt.Println(h[idx], "=", r)
			m[h[idx]] = r
		}
		records = append(records, m)
		fmt.Println("--------------------------------------------------")
	}
	return records
}
