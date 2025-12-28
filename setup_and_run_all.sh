#!/bin/bash
set -e

# تعريف الألوان
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🚀 البدء في إعداد المشروع وتثبيت العقد الذكي للشهادات...${NC}"
echo "=================================================="

# 1. إعداد المسارات (Environment Path)
export PATH=${PWD}/bin:$PATH
export FABRIC_CFG_PATH=${PWD}/config/

# 2. تنظيف وإعادة تشغيل شبكة Fabric
echo -e "${GREEN}🌐 الخطوة 1: إعادة تشغيل الشبكة...${NC}"
cd test-network
./network.sh down
./network.sh up createChannel -c mychannel -ca
cd ..

# 3. تحديث مكتبات Go وتصحيح العقد الذكي
echo -e "${GREEN}📦 الخطوة 2: تحديث مكتبات العقد الذكي (Go)...${NC}"
pushd asset-transfer-basic/chaincode-go
# هذا الأمر يحل مشكلة الـ Undefined ويحمل المكتبات المطلوبة
go mod tidy
popd

# 4. نشر العقد الذكي (Deploy)
echo -e "${GREEN}📜 الخطوة 3: نشر العقد الذكي للشهادات...${NC}"
cd test-network
# استخدام المسار الدقيق كما يظهر في صورك
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
cd ..

# 5. تشغيل اختبارات Caliper
echo -e "${GREEN}⚡ الخطوة 4: تشغيل اختبار الأداء (Caliper)...${NC}"
cd caliper-workspace

# التحقق من وجود المفتاح الخاص أوتوماتيكياً
echo "🔑 البحث عن المفتاح الخاص للـ Admin..."
KEY_DIR="../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore"
PVT_KEY=$(ls $KEY_DIR/*_sk)

# إنشاء ملف إعدادات الشبكة
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

# تنفيذ الاختبار
npx caliper launch manager \
    --caliper-workspace . \
    --caliper-networkconfig networks/networkConfig.yaml \
    --caliper-benchconfig benchmarks/benchConfig.yaml \
    --caliper-flow-only-test

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}🎉 تم الانتهاء بنجاح! راجع ملف report.html للنتائج.${NC}"
