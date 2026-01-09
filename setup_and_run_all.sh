#!/bin/bash
set -e

# تعريف الألوان
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🚀 البدء في إعداد مشروع SecureBlockCert (محاكاة دراسة عمر سعد - AES)...${NC}"
echo "=================================================="

# 1. التأكد من وجود الأدوات (Fabric v2.5.9)
if [ ! -d "bin" ]; then
    echo "⬇️ Downloading Fabric binaries and Docker images..."
    curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.5.9 1.5.7
else
    echo "✅ Fabric tools found."
fi

export PATH=${PWD}/bin:$PATH
export FABRIC_CFG_PATH=${PWD}/config/

# 2. تنظيف وإعادة تشغيل شبكة Fabric
echo -e "${GREEN}🌐 الخطوة 1: إعادة تشغيل الشبكة...${NC}"
cd test-network
./network.sh down
./network.sh up createChannel -c mychannel -ca
cd ..

# 3. تحديث مكتبات Go وتجهيز التشفير (AES)
echo -e "${GREEN}📦 الخطوة 2: تجهيز العقد الذكي بخوارزمية AES لحماية الخصوصية...${NC}"
pushd asset-transfer-basic/chaincode-go
# تهيئة الموديول وعمل Vendor لضمان وجود المكتبات داخل الحاوية
go mod tidy
go mod vendor
popd

# 4. نشر العقد الذكي
echo -e "${GREEN}📜 الخطوة 3: نشر العقد الذكي (Secure Chaincode)...${NC}"
cd test-network
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
cd ..

# 5. تهيئة Caliper
echo -e "${GREEN}⚙️ الخطوة 4: تهيئة Caliper وربط النسخة المستقرة...${NC}"
cd caliper-workspace
if [ ! -d "node_modules" ]; then
    npm install
fi
# الربط بـ 2.4 هو الأفضل توافقاً مع Fabric 2.5
npx caliper bind --caliper-bind-sut fabric:2.2

# 6. تحديث ملف إعدادات الشبكة (إصلاح التنسيق)
echo "🔑 البحث عن المفتاح الخاص للـ Admin..."
KEY_DIR="../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore"
PVT_KEY=$(ls $KEY_DIR/*_sk | head -n 1)

echo "⚙️ Generating Clean Network Config..."
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

# 7. تنفيذ الاختبار المطور (AES + HMAC)
echo -e "${GREEN}🚀 تشغيل اختبار Caliper (Issue & Verify)...${NC}"
npx caliper launch manager \
    --caliper-workspace . \
    --caliper-networkconfig networks/networkConfig.yaml \
    --caliper-benchconfig benchmarks/benchConfig.yaml \
    --caliper-flow-only-test

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}🎉 تم الانتهاء بنجاح!${NC}"
echo -e "${GREEN}📄 التقرير متوفر في: caliper-workspace/report.html${NC}"
echo -e "${GREEN}💡 قارن Latency في جولة Verify لتلاحظ تأثير الأمان المضاف.${NC}"
