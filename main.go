package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/VictoriaMetrics/metrics"
	"log"
	"os"
	"time"
)

//TODO add switch to chain certs or only domain

func readDomains(path string) []string {
	var domainsList []string
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		domainsList = append(domainsList, scanner.Text())
	}

	return domainsList
}

func testDomain(domain string, chain bool) {
	conn, err := tls.Dial("tcp", domain+":443", nil)
	if err != nil {
		log.Println("Error in Dial", err)
		//TODO add counter of failed domains
		return
	}
	defer conn.Close()
	err = conn.VerifyHostname(domain)
	if err != nil {
		panic("Hostname doesn't match with certificate: " + err.Error())
	}
	certs := conn.ConnectionState().PeerCertificates
	for _, cert := range certs {
		log.Printf("Processing cert: %s", cert.Subject.CommonName)
		notAfter := cert.NotAfter
		diffDays := int(notAfter.Sub(time.Now()).Hours() / 24)

		// metric construct
		name := fmt.Sprintf(`ssl_expiration_days_left{subject_cn=%q, issuer=%q, pka=%q, serial="%d"}`,
			cert.Subject.CommonName,
			cert.Issuer.CommonName,
			cert.PublicKeyAlgorithm,
			cert.SerialNumber)
		metrics.GetOrCreateGauge(name, func() float64 {
			return float64(diffDays)
		})
		if !chain {
			break
		}
	}
}

func main() {
	domains := readDomains("./domains")
	for _, domain := range domains {
		log.Printf("Processing domain: %s", domain)
		testDomain(domain, false)
	}

	metrics.WritePrometheus(os.Stdout, false)
}
