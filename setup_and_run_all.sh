#!/bin/bash
set -e

# تعريف الألوان
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🚀 البدء في إعداد مشروع SecureBlockCert (محاكاة دراسة عمر سعد - AES)...${NC}"
echo "=================================================="

# 1. التأكد من وجود الأدوات
if [ ! -d "bin" ]; then
    echo "⬇️ Downloading Fabric binaries..."
    curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.5.9 1.5.7
else
    echo "✅ Fabric tools found."
fi

export PATH=${PWD}/bin:$PATH
export FABRIC_CFG_PATH=${PWD}/config/

# 2. إعادة تشغيل الشبكة
echo -e "${GREEN}🌐 الخطوة 1: إعادة تشغيل الشبكة...${NC}"
cd test-network
./network.sh down
./network.sh up createChannel -c mychannel -ca
cd ..

# 3. تجهيز العقد الذكي (هنا تم دمج الكود والإصلاحات)
echo -e "${GREEN}📦 الخطوة 2: كتابة العقد الذكي (AES) وتنظيف المجلدات...${NC}"
pushd asset-transfer-basic/chaincode-go

# --- [بداية الإصلاحات المدمجة] ---

# أ) حذف الملفات والمجلدات المتعارضة
echo "🗑️ تنظيف الملفات القديمة (chaincode folder, assetTransfer.go)..."
rm -rf chaincode       # حذف المجلد الفرعي المسبب للمشكلة
rm -f assetTransfer.go # حذف الملف القديم إن وجد

# ب) كتابة ملف main.go الجديد مباشرة من هنا
echo "✍️ كتابة كود main.go الجديد (AES Encryption)..."
cat << 'EOF' > main.go
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract defines the structure for our chaincode
type SmartContract struct {
	contractapi.Contract
}

var encryptionKey = []byte("asupersecretkeythatis32byteslong")

type Certificate struct {
	ID          string `json:"ID"`
	StudentName string `json:"StudentName"`
	Degree      string `json:"Degree"`
	IssueDate   string `json:"IssueDate"`
	Issuer      string `json:"Issuer"`
}

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

func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return certJSON != nil, nil
}

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

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}
EOF

# ج) تحديث المكتبات
echo "🔄 تحديث المكتبات (go mod tidy & vendor)..."
rm -f go.sum
rm -rf vendor
go mod tidy
go mod vendor

# --- [نهاية الإصلاحات] ---
popd

# 4. نشر العقد الذكي
echo -e "${GREEN}📜 الخطوة 3: نشر العقد الذكي...${NC}"
cd test-network
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
cd ..

# 5. تهيئة Caliper
echo -e "${GREEN}⚙️ الخطوة 4: تهيئة Caliper...${NC}"
cd caliper-workspace
if [ ! -d "node_modules" ]; then
    npm install
fi
npx caliper bind --caliper-bind-sut fabric:2.2

# 6. إعداد ملف الشبكة للمفاتيح
echo "🔑 تحديث مفاتيح Admin..."
KEY_DIR="../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore"
PVT_KEY=$(ls $KEY_DIR/*_sk | head -n 1)

if [ -z "$PVT_KEY" ]; then
    echo -e "${RED}❌ خطأ: لم يتم العثور على المفتاح الخاص!${NC}"
    exit 1
fi

cat << EOF > networks/networkConfig.yaml
name: Caliper-Fabric
version: "2.0.0"
caliper:
  blockchain: fabric
channels:
  - channelName: mychannel
    contracts:
      - id: basic
organizations:
  - mspid: Org1MSP
    identities:
      certificates:
        - name: 'User1'
          clientPrivateKey:
            path: '$PVT_KEY'
          clientSignedCert:
            path: '../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/cert.pem'
    connectionProfile:
      path: '../test-network/organizations/peerOrganizations/org1.example.com/connection-org1.yaml'
      discover: true
EOF

# 7. التشغيل النهائي
echo -e "${GREEN}🚀 تشغيل اختبار Caliper...${NC}"
npx caliper launch manager \
    --caliper-workspace . \
    --caliper-networkconfig networks/networkConfig.yaml \
    --caliper-benchconfig benchmarks/benchConfig.yaml \
    --caliper-flow-only-test

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}🎉 تم الانتهاء بنجاح! التقرير جاهز.${NC}"
