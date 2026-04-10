package main

import (
	"fmt"
	"log"
	"net/http"
)

func mockSOAP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	
	payload := `<?xml version="1.0"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
  <SOAP-ENV:Body>
    <m:GetAccountResponse xmlns:m="http://crom.sec/java">
      <m:TenantID>JAVA_SPRING_MASSIVE_XML_PAYLOAD</m:TenantID>
    </m:GetAccountResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	w.Write([]byte(payload))
}

func main() {
	fmt.Println(" [JAVA-SPRING-MOCK] Rodando na porta 8082")
	http.HandleFunc("/soap", mockSOAP)
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("Fail: %v", err)
	}
}
