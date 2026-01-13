package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log" // تم إضافتها لاستخدامها في دالة main

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract defines the structure for our chaincode
type SmartContract struct {
	contractapi.Contract
}

// encryptionKey: مفتاح التشفير (32 بايت)
var encryptionKey = []byte("asupersecretkeythatis32byteslong")

// Certificate: هيكل الشهادة الأكاديمية
type Certificate struct {
	ID          string `json:"ID"`
	StudentName string `json:"StudentName"`
	Degree      string `json:"Degree"`
	IssueDate   string `json:"IssueDate"`
	Issuer      string `json:"Issuer"`
}

// =========================================================================================
// دوال الـ Chaincode الرئيسية (Transaction Functions)
// =========================================================================================

// IssueCertificate: إصدار شهادة وتشفيرها
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, name string, degree string, date string, issuer string) error {
	
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the certificate %s already exists", id)
	}

	cert := Certificate{
		ID:          id,
		StudentName: name,
		Degree:      degree,
		IssueDate:   date,
		Issuer:      issuer,
	}

	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	encryptedData, err := encrypt(certJSON, encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %v", err)
	}

	return ctx.GetStub().PutState(id, []byte(encryptedData))
}

// ReadCertificate: قراءة وفك تشفير الشهادة
func (s *SmartContract) ReadCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	
	encryptedDataBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if encryptedDataBytes == nil {
		return nil, fmt.Errorf("the certificate %s does not exist", id)
	}

	decryptedJSON, err := decrypt(string(encryptedDataBytes), encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %v", err)
	}

	var cert Certificate
	err = json.Unmarshal([]byte(decryptedJSON), &cert)
	if err != nil {
		return nil, err
	}

	return &cert, nil
}

// CertificateExists: التحقق من الوجود
func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return certJSON != nil, nil
}

// =========================================================================================
// وظائف التشفير المساعدة
// =========================================================================================

func encrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(cryptoText string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// =========================================================================================
// دالة MAIN (الضرورية لتشغيل الـ Chaincode)
// =========================================================================================
func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}
