package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var gcpInputHeader = `First Name [Required],Last Name [Required],Email Address [Required],Password [Required],Password Hash Function [UPLOAD ONLY],Org Unit Path [Required],New Primary Email [UPLOAD ONLY],Status [READ ONLY],Last Sign In [READ ONLY],Recovery Email,Home Secondary Email,Work Secondary Email,Recovery Phone [MUST BE IN THE E.164 FORMAT],Work Phone,Home Phone,Mobile Phone,Work Address,Home Address,Employee ID,Employee Type,Employee Title,Manager Email,Department,Cost Center,2sv Enrolled [READ ONLY],2sv Enforced [READ ONLY],Building ID,Floor Name,Floor Section,Email Usage [READ ONLY],Drive Usage [READ ONLY],Total Storage [READ ONLY],Change Password at Next Sign-In,New Status [UPLOAD ONLY]`
var azureInputHeader = `userPrincipalName,displayName,surname,mail,givenName,objectId,userType,jobTitle,department,accountEnabled,usageLocation,streetAddress,state,country,physicalDeliveryOfficeName,city,postalCode,telephoneNumber,mobile,authenticationPhoneNumber,authenticationAlternativePhoneNumber,authenticationEmail,alternateEmailAddress,ageGroup,consentProvidedForMinor,legalAgeGroupClassification`
var azureOutputHeader = `Name [displayName] Required,User name [userPrincipalName] Required,Initial password [passwordProfile] Required,Block sign in (Yes/No) [accountEnabled] Required,First name [givenName],Last name [surname],Job title [jobTitle],Department [department],Usage location [usageLocation],Street address [streetAddress],State or province [state],Country or region [country],Office [physicalDeliveryOfficeName],City [city],ZIP or postal code [postalCode],Office phone [telephoneNumber],Mobile phone [mobile]`

func check(ptr *string, msg string) bool {
	//fmt.Println(*ptr)
	if *ptr == "" {
		fmt.Println(msg)
		return false
	}
	return true
}

var verbose *bool

func main() {
	var source, target, sourceFile, targetFile string

	flag.StringVar(&source, "source", "gcp", "Source IdP")
	flag.StringVar(&target, "target", "azure", "Target IdP")
	flag.StringVar(&sourceFile, "sourceFile", "", "CSV of source IdP")
	flag.StringVar(&targetFile, "targetFile", "", "CSV of target IdP")

	verbose = flag.Bool("verbose", false, "Log to console")

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
		mappingFile := "mapping/" + source + "_to_" + target + ".txt"
		writeAzureUser(sourceRecords, mappingFile, targetFile)
	} else if target == "aws" {

	}
}

func writeAzureUser(sourceRecords []map[string]string, mappingFile string, targetFile string) {
	mf, err := os.Open(mappingFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer mf.Close()

	//Create header mapping between two formats
	headerMap := make(map[string]string)
	scanner := bufio.NewScanner(mf)
	for scanner.Scan() {
		line := scanner.Text()
		ss := strings.Split(line, "=")
		if len(ss) > 1 && len(ss[0]) > 0 && len(ss[1]) > 0 {
			headerMap[ss[0]] = ss[1]
		}
	}

	//Display mappings
	for k, v := range headerMap {
		logToConsole("mapping", k, "to", v)
	}

	logToConsole()
	//Create new records
	var newRecords []map[string]string
	azureHeaders := strings.Split(azureOutputHeader, ",")
	for _, sourceRerord := range sourceRecords {
		newRecord := make(map[string]string)

		for _, azureHeader := range azureHeaders {
			logToConsole("azureHeader:", azureHeader)
			sourceHeader := headerMap[azureHeader]
			logToConsole("sourceHeader:", sourceHeader)
			if sourceHeader != "" {
				if sourceHeader == "No" || sourceHeader == "Yes" {
					newRecord[azureHeader] = sourceHeader
				} else {
					//There could be multiple source headers
					var buffer bytes.Buffer
					for _, h := range strings.Split(sourceHeader, "+") {
						header := strings.TrimSpace(h)
						value := sourceRerord[header]
						logToConsole("\theader:", header, ",v:", value)
						if value != "" {
							buffer.WriteString(value)
							buffer.WriteString(" ")
						}
					}
					newRecord[azureHeader] = strings.TrimSpace(buffer.String())
				}
			}

			if strings.HasSuffix(azureHeader, "Required") && newRecord[azureHeader] == "" {
				// A required field but no value. Use default
				newRecord[azureHeader] = "default-todo"
			}
		}

		newRecords = append(newRecords, newRecord)
	}

	//Display new records
	logToConsole("result: ")
	logToConsole("==newRecords================================================")
	for _, newRecord := range newRecords {
		for k, v := range newRecord {
			logToConsole("\t", k, "=", v)
		}
		logToConsole("\t----------")
	}

	fmt.Println("Generated ", len(newRecords), " records")

	//Write to file
	of, err := os.Create(targetFile)
	if err != nil {
		fmt.Println(err)
		of.Close()
		return
	}

	//Write header first
	fmt.Fprintln(of, "version:v1.0\r")
	fmt.Fprintln(of, azureOutputHeader+"\r")

	for _, newRecord := range newRecords {
		var buffer bytes.Buffer
		for _, header := range azureHeaders {
			value := newRecord[header]
			if strings.TrimSpace(value) != "" {
				logToConsole(header, "=", value)
				buffer.WriteString(value)
			}
			buffer.WriteString(",")
		}

		line := buffer.String()
		//Remove last ,
		line = line[:len(line)-1]
		logToConsole(line)
		fmt.Fprintln(of, line+"\r")
	}
	err = of.Close()
}

func readGcpUsr(gcpCsv string) []map[string]string {
	file, err := os.Open(gcpCsv)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()

	h := strings.Split(gcpInputHeader, ",")

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
		if lineNumber == 1 {
			continue
		}

		m := make(map[string]string)
		for idx, field := range record {
			if field == "" {
				continue
			}
			logToConsole(h[idx], "=", field)
			m[h[idx]] = field
		}
		records = append(records, m)
		logToConsole("--------------------------------------------------")
	}

	fmt.Println("Found ", len(records), " rows")
	return records
}

func logToConsole(line ...string) {
	if *verbose {
		var buffer bytes.Buffer
		for _, l := range line {
			buffer.WriteString(l)
		}
		fmt.Println(buffer.String())
	}
}
