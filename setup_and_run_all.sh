#!/bin/bash
set -e

# تعريف الألوان
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🚀 البدء في تشغيل مشروع SecureBlockCert (النسخة الذكية)...${NC}"
echo "=================================================="

# 1. إعداد الأدوات
if [ ! -d "bin" ]; then
    echo "⬇️ Downloading Fabric binaries..."
    curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.5.9 1.5.7
fi
export PATH=${PWD}/bin:$PATH
export FABRIC_CFG_PATH=${PWD}/config/

# 2. إعادة تشغيل الشبكة
echo -e "${GREEN}🌐 الخطوة 1: إعادة تشغيل الشبكة...${NC}"
cd test-network
./network.sh down
./network.sh up createChannel -c mychannel -ca
cd ..

# ---------------------------------------------------------
# 3. العقد الذكي (Go)
# ---------------------------------------------------------
echo -e "${GREEN}📦 الخطوة 2: تجهيز العقد الذكي (AES)...${NC}"
pushd asset-transfer-basic/chaincode-go
rm -rf chaincode assetTransfer.go main.go

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
	if err != nil { return err }
	if exists { return fmt.Errorf("the certificate %s already exists", id) }

	cert := Certificate{ID: id, StudentName: name, Degree: degree, IssueDate: date, Issuer: issuer}
	certJSON, err := json.Marshal(cert)
	if err != nil { return err }

	encryptedData, err := encrypt(certJSON, encryptionKey)
	if err != nil { return fmt.Errorf("failed to encrypt: %v", err) }

	return ctx.GetStub().PutState(id, []byte(encryptedData))
}

func (s *SmartContract) ReadCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	encryptedDataBytes, err := ctx.GetStub().GetState(id)
	if err != nil { return nil, fmt.Errorf("failed to read: %v", err) }
	if encryptedDataBytes == nil { return nil, fmt.Errorf("the certificate %s does not exist", id) }

	decryptedJSON, err := decrypt(string(encryptedDataBytes), encryptionKey)
	if err != nil { return nil, fmt.Errorf("failed to decrypt: %v", err) }

	var cert Certificate
	err = json.Unmarshal([]byte(decryptedJSON), &cert)
	if err != nil { return nil, err }
	return &cert, nil
}

func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil { return false, fmt.Errorf("failed to read: %v", err) }
	return certJSON != nil, nil
}

func encrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil { return "", err }
	gcm, err := cipher.NewGCM(block)
	if err != nil { return "", err }
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil { return "", err }
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(cryptoText string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil { return "", err }
	block, err := aes.NewCipher(key)
	if err != nil { return "", err }
	gcm, err := cipher.NewGCM(block)
	if err != nil { return "", err }
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize { return "", fmt.Errorf("ciphertext too short") }
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil { return "", err }
	return string(plaintext), nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil { log.Panicf("Error creating chaincode: %v", err) }
	if err := chaincode.Start(); err != nil { log.Panicf("Error starting chaincode: %v", err) }
}
EOF

echo "🔄 تحديث المكتبات..."
rm -f go.sum vendor -rf
go mod tidy
go mod vendor
popd

# ---------------------------------------------------------
# 4. نشر العقد الذكي
# ---------------------------------------------------------
echo -e "${GREEN}📜 الخطوة 3: نشر العقد الذكي...${NC}"
cd test-network
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
cd ..

# ---------------------------------------------------------
# 5. تكوين Caliper (هنا الحل الذكي)
# ---------------------------------------------------------
echo -e "${GREEN}⚙️ الخطوة 4: إعداد ملفات Caliper الذكية...${NC}"
cd caliper-workspace
mkdir -p workload benchmarks

# A. ملف الإصدار (Issue)
cat << 'EOF' > workload/issueCertificate.js
'use strict';
const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class IssueCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }
    async submitTransaction() {
        this.txIndex++;
        // الترقيم يبدأ من 1 دائماً
        const certID = `CERT_${this.workerIndex}_${this.txIndex}`;
        const studentName = `Student_${this.workerIndex}_${this.txIndex}`;
        const request = {
            contractId: 'basic',
            contractFunction: 'IssueCertificate',
            contractArguments: [certID, studentName, 'PhD', new Date().toISOString(), 'UUM'],
            readOnly: false
        };
        await this.sutAdapter.sendRequests(request);
    }
}
function createWorkloadModule() { return new IssueCertificateWorkload(); }
module.exports.createWorkloadModule = createWorkloadModule;
EOF

# B. ملف التحقق (Verify) - الحل الذكي للدوران
cat << 'EOF' > workload/verifyCertificate.js
'use strict';
const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class VerifyCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }
    async submitTransaction() {
        this.txIndex++;
        
        // ---[ الحل الذكي ]---
        // بدلاً من طلب رقم جديد قد لا يكون موجوداً، نستخدم المعامل % (Modulo)
        // لنضمن أننا نطلب دائماً رقماً بين 1 و 50 (التي تم إنشاؤها بالتأكيد)
        // هذا يمنع خطأ "does not exist"
        const safeIndex = (this.txIndex % 50) + 1;
        
        const certID = `CERT_${this.workerIndex}_${safeIndex}`;

        const request = {
            contractId: 'basic',
            contractFunction: 'ReadCertificate',
            contractArguments: [certID],
            readOnly: true
        };
        await this.sutAdapter.sendRequests(request);
    }
}
function createWorkloadModule() { return new VerifyCertificateWorkload(); }
module.exports.createWorkloadModule = createWorkloadModule;
EOF

# C. ملف البنش مارك (Benchmark Config)
# نضمن هنا أننا ننشئ 60 معاملة، ونقرأ 100 مرة (القراءة ستكون آمنة الآن)
cat << EOF > benchmarks/benchConfig.yaml
test:
  name: certificate-benchmark
  description: SecureBlockCert Benchmark
  workers:
    type: local
    number: 2
  rounds:
    - label: Issue Phase
      description: Create certificates
      txNumber: 60  # ننشئ 60 شهادة لكل عامل
      rateControl:
        type: fixed-rate
        opts:
          tps: 10
      workload:
        module: workload/issueCertificate.js
    
    - label: Verify Phase
      description: Read certificates (Decryption)
      txNumber: 100 # نقرأ 100 مرة (سيتم تكرار قراءة الـ 60 شهادة الموجودة)
      rateControl:
        type: fixed-rate
        opts:
          tps: 20
      workload:
        module: workload/verifyCertificate.js
EOF

# تثبيت وربط
if [ ! -d "node_modules" ]; then npm install; fi
npx caliper bind --caliper-bind-sut fabric:2.2

# ---------------------------------------------------------
# 6. تشغيل الاختبار
# ---------------------------------------------------------
echo "🔑 تحديث المفاتيح..."
KEY_DIR="../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore"
PVT_KEY=$(ls $KEY_DIR/*_sk | head -n 1)

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

echo -e "${GREEN}🚀 تشغيل Caliper (Issue & Smart Verify)...${NC}"
npx caliper launch manager \
    --caliper-workspace . \
    --caliper-networkconfig networks/networkConfig.yaml \
    --caliper-benchconfig benchmarks/benchConfig.yaml \
    --caliper-flow-only-test

echo -e "${GREEN}🎉 تم الانتهاء! يجب أن يختفي خطأ 'does not exist' الآن.${NC}"
