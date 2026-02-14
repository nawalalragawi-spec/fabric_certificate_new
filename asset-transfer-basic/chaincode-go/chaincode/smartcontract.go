package chaincode

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Certificate struct {
	ID        string `json:"ID"`
	CertHash  string `json:"CertHash"`
	IssueDate string `json:"IssueDate"`
	Issuer    string `json:"Issuer"`
	Owner     string `json:"Owner"`
	IsLocked  bool   `json:"IsLocked"`
}

const AES_KEY = "12345678901234567890123456789012"

func decrypt(encryptedData string) (string, error) {
	parts := strings.Split(encryptedData, ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid format")
	}
	iv, _ := hex.DecodeString(parts[0])
	ciphertext, _ := hex.DecodeString(parts[1])
	authTag, _ := hex.DecodeString(parts[2])
	block, err := aes.NewCipher([]byte(AES_KEY))
	if err != nil { return "", err }
	aesgcm, err := cipher.NewGCM(block)
	if err != nil { return "", err }
	fullCiphertext := append(ciphertext, authTag...)
	plaintext, err := aesgcm.Open(nil, iv, fullCiphertext, nil)
	if err != nil { return "", err }
	return string(plaintext), nil
}

func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, owner string, issuer string, issueDate string, encryptedHash string) error {
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "admin")
	if err != nil { return fmt.Errorf("unauthorized") }
	exists, err := s.CertificateExists(ctx, id)
	if err != nil { return err }
	if exists { return fmt.Errorf("exists") }
	certificate := Certificate{
		ID: id, Owner: owner, Issuer: issuer, IssueDate: issueDate, CertHash: encryptedHash, IsLocked: true,
	}
	certBytes, err := json.Marshal(certificate)
	if err != nil { return err }
	return ctx.GetStub().PutState(id, certBytes)
}

func (s *SmartContract) LockCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil { return err }
	clientID, _ := ctx.GetClientIdentity().GetID()
	if certificate.Owner != clientID { return fmt.Errorf("unauthorized") }
	certificate.IsLocked = true
	certBytes, err := json.Marshal(certificate)
	if err != nil { return err }
	return ctx.GetStub().PutState(id, certBytes)
}

func (s *SmartContract) UnlockCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil { return err }
	clientID, _ := ctx.GetClientIdentity().GetID()
	if certificate.Owner != clientID { return fmt.Errorf("unauthorized") }
	certificate.IsLocked = false
	certBytes, err := json.Marshal(certificate)
	if err != nil { return err }
	return ctx.GetStub().PutState(id, certBytes)
}

func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string, currentHash string) (bool, error) {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil { return false, err }
	if certificate.IsLocked { return false, fmt.Errorf("locked") }
	decryptedStoredHash, err := decrypt(certificate.CertHash)
	if err != nil { return false, err }
	return decryptedStoredHash == currentHash, nil
}

func (s *SmartContract) ReadCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	certBytes, err := ctx.GetStub().GetState(id)
	if err != nil { return nil, err }
	if certBytes == nil { return nil, fmt.Errorf("not found") }
	var certificate Certificate
	err = json.Unmarshal(certBytes, &certificate)
	if err != nil { return nil, err }
	return &certificate, nil
}

func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certBytes, err := ctx.GetStub().GetState(id)
	if err != nil { return false, err }
	return certBytes != nil, nil
}

func (s *SmartContract) GetAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Certificate, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil { return nil, err }
	defer resultsIterator.Close()
	var certificates []*Certificate
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil { return nil, err }
		var certificate Certificate
		err = json.Unmarshal(queryResponse.Value, &certificate)
		if err != nil { return nil, err }
		certificates = append(certificates, &certificate)
	}
	return certificates, nil
}
