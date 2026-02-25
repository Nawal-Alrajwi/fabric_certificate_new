#!/bin/bash
set -e

# 1. تنظيف عميق للبيئة لضمان عدم تداخل الإعدادات القديمة
docker rm -f $(docker ps -aq) || true
docker volume prune -f
echo "🧹 Performing deep-clean for old Docker images..."
DEV_IMAGE_IDS=$(docker images --format '{{.Repository}} {{.ID}}' | awk '$1 ~ /^(dev-|dev-peer)/ {print $2}' || true)
if [ -n "$DEV_IMAGE_IDS" ]; then
  docker rmi -f $DEV_IMAGE_IDS || true
fi

# 2. مسح التقارير القديمة وتجهيز الـ Workspace
rm -f caliper-workspace/report.html
cd caliper-workspace && rm -rf networks/networkConfig.yaml && cd ..

GREEN='\033[0;32m'
NC='\033[0m'
echo -e "${GREEN}🚀 Starting Full Project Setup (Fabric + Caliper)...${NC}"

# 3. تشغيل الشبكة مع تفعيل CouchDB (ضروري لأداء ورقة 2025)
cd test-network
./network.sh down
./network.sh up createChannel -c mychannel -ca -s couchdb
# إعطاء وقت لـ CouchDB للاستقرار وتجنب خطأ connection refused
sleep 15 
cd ..

# 4. نشر العقد الذكي بسياسة OR (لحل فشل الحذف وضمان استقرار الصلاحيات)
echo "📜 Deploying Smart Contract with OR Policy..."
cd test-network
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go -ccep "OR('Org1MSP.peer','Org2MSP.peer')"
cd ..

# 5. إعداد Caliper واكتشاف المفاتيح آلياً
cd caliper-workspace
if [ ! -d "node_modules" ]; then
  npm install
  npx caliper bind --caliper-bind-sut fabric:2.2
fi

echo "🔑 Detecting Private Keys for Both Orgs..."
KEY_DIR1="../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore"
PVT_KEY1=$(ls $KEY_DIR1/*_sk)
KEY_DIR2="../test-network/organizations/peerOrganizations/org2.example.com/users/User1@org2.example.com/msp/keystore"
PVT_KEY2=$(ls $KEY_DIR2/*_sk)

# 6. توليد ملف networkConfig.yaml بتنسيق YAML صحيح ومسارات دقيقة
echo "⚙️ Generating correct network config..."
mkdir -p networks
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
        - name: 'User1@org1.example.com'
          clientPrivateKey:
            path: '$PVT_KEY1'
          clientSignedCert:
            path: '../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem'
    connectionProfile:
      path: '../test-network/organizations/peerOrganizations/org1.example.com/connection-org1.yaml'
      discover: true

  - mspid: Org2MSP
    identities:
      certificates:
        - name: 'User1@org2.example.com'
          clientPrivateKey:
            path: '$PVT_KEY2'
          clientSignedCert:
            path: '../test-network/organizations/peerOrganizations/org2.example.com/users/User1@org2.example.com/msp/signcerts/User1@org2.example.com-cert.pem'
    connectionProfile:
      path: '../test-network/organizations/peerOrganizations/org2.example.com/connection-org2.yaml'
      discover: true
EOF

# 7. تشغيل الاختبار النهائي
echo "🔥 Running Benchmark..."
npx caliper launch manager \
    --caliper-workspace . \
    --caliper-networkconfig networks/networkConfig.yaml \
    --caliper-benchconfig benchmarks/benchConfig.yaml \
    --caliper-flow-only-test

echo "✅ Finished. Report generated at caliper-workspace/report.html"
