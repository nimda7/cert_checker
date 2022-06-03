package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/VictoriaMetrics/metrics"
	"log"
	"os"
	"strconv"
	"time"
)

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
		errDomainCounter := fmt.Sprintf(`error_domains{fqdn="%s"}`, domain)
		metrics.GetOrCreateCounter(errDomainCounter).Add(1)
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

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func main() {
	checkCertChain, _ := strconv.ParseBool(getEnv("CS_CHAIN", "false"))
	domainsPath := getEnv("CS_DOMAINS", "./domains")

	log.Printf("Processing domains list from: %s", domainsPath)
	domains := readDomains(domainsPath)
	for _, domain := range domains {
		log.Printf("Processing domain: %s", domain)
		testDomain(domain, checkCertChain)
	}

	metrics.WritePrometheus(os.Stdout, false)
}
