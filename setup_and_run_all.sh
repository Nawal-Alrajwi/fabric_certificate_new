#!/bin/bash
set -e

# تعريف الألوان
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🚀 البدء في إعداد المشروع وتثبيت العقد الذكي المطور (SHA-3)...${NC}"
echo "=================================================="

if [ ! -d "bin" ]; then
    echo "⬇️ Downloading Fabric binaries and Docker images (v2.5.9)..."
    curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.5.9 1.5.7
else
    echo "✅ Fabric tools found. Pulling/Verifying Docker images..."
    curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.5.9 1.5.7 -s -b
fi

# 1. إعداد المسارات
export PATH=${PWD}/bin:$PATH
export FABRIC_CFG_PATH=${PWD}/config/

# 2. تنظيف وإعادة تشغيل شبكة Fabric
echo -e "${GREEN}🌐 الخطوة 1: إعادة تشغيل الشبكة...${NC}"
cd test-network
./network.sh down
./network.sh up createChannel -c mychannel -ca
cd ..

# 3. تحديث مكتبات Go وتثبيت مكتبة SHA-3 (التعديل الجوهري هنا)
echo -e "${GREEN}📦 الخطوة 2: تحديث مكتبات العقد الذكي وإضافة SHA-3...${NC}"
pushd asset-transfer-basic/chaincode-go
# تهيئة الموديول والتأكد من جلب مكتبة التشفير الجديدة
go mod tidy
go get golang.org/x/crypto/sha3
popd

# 4. نشر العقد الذكي
echo -e "${GREEN}📜 الخطوة 3: نشر العقد الذكي المطور...${NC}"
cd test-network
# ملاحظة: تأكد أن اسم العقد 'basic' يطابق ما هو في ملف الإعدادات
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
cd ..

# 5. تهيئة Caliper
echo -e "${GREEN}⚙️ الخطوة 4: تهيئة Caliper وربط المكتبات...${NC}"
cd caliper-workspace
# تثبيت التبعيات إذا لم تكن موجودة
if [ ! -d "node_modules" ]; then
    npm install
fi
npx caliper bind --caliper-bind-sut fabric:2.2

# 6. تحديث ملف إعدادات الشبكة (Network Config)
echo "🔑 البحث عن المفتاح الخاص للـ Admin..."
KEY_DIR="../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore"
PVT_KEY=$(ls $KEY_DIR/*_sk)

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

# 7. تنفيذ الاختبار المطور
echo -e "${GREEN}🚀 تشغيل اختبار Caliper المطور...${NC}"
npx caliper launch manager \
    --caliper-workspace . \
    --caliper-networkconfig networks/networkConfig.yaml \
    --caliper-benchconfig benchmarks/benchConfig.yaml \
    --caliper-flow-only-test

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}🎉 تم الانتهاء! ستلاحظ الآن تحسناً كبيراً في نتائج VerifyCertificate.${NC}"
