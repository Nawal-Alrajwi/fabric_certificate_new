#!/bin/bash
set -e
# تشغيل سكربت إصلاح الصلاحيات فقط في بيئة CI أو عند طلب صريح عبر FIX_PERMISSIONS
# يمكن فرض التشغيل محليًا بتشغيل: FIX_PERMISSIONS=true ./setup_and_run_all.sh
if [ "${CI:-}" = "true" ] || [ "${CI:-}" = "1" ] || [ -n "${GITHUB_ACTIONS:-}" ] || [ "${FIX_PERMISSIONS:-}" = "true" ]; then
  if [ -x "./scripts/fix-permissions.sh" ]; then
    echo "🔐 Running scripts/fix-permissions.sh to fix permissions (CI or FIX_PERMISSIONS set)..."
    ./scripts/fix-permissions.sh || true
  else
    echo "⚠️ scripts/fix-permissions.sh not found or not executable. Skipping."
  fi
else
  echo "ℹ️ Not in CI and FIX_PERMISSIONS not set; skipping permission fix."
fi
# 1. مسح أي حاويات أو شبكات قديمة متبقية بالقوة
docker rm -f $(docker ps -aq) || true
docker volume prune -f

# --------------------------------------------------------
# Deep Clean: إزالة صور Docker التي تبدأ بـ dev-* أو dev-peer*
# هذا يضمن بناء صور العقد الذكي الجديدة بدلاً من إعادة استخدام القديمة
# --------------------------------------------------------
echo -e "\n🧹 Performing deep-clean for Docker images starting with dev-*..."
# جمع معرفات الصور المطابقة
DEV_IMAGE_IDS=$(docker images --format '{{.Repository}} {{.ID}}' | awk '$1 ~ /^(dev-|dev-peer)/ {print $2}' || true)
if [ -n "$DEV_IMAGE_IDS" ]; then
  echo "Found dev images: $DEV_IMAGE_IDS"
  docker rmi -f $DEV_IMAGE_IDS || true
else
  echo "No dev-* images found."
fi

# 2. مسح التقارير القديمة للتأكد أن التقرير الناتج هو الجديد
rm -f caliper-workspace/report.html

# 3. التأكد من تحديث الـ Workspace
cd caliper-workspace && rm -rf networks/networkConfig.yaml && cd ..
# تعريف الألوان للنصوص
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'
echo -e "nawal2"

echo -e "${GREEN}🚀 Starting Full Project Setup (Fabric + Caliper)...${NC}"
echo "=================================================="

# --------------------------------------------------------
# 1. التأكد من وجود الأدوات
# --------------------------------------------------------
echo -e "${GREEN}📦 Step 1: Checking Fabric Binaries...${NC}"
if [ ! -d "bin" ]; then
    echo "⬇️ Downloading Fabric tools..."
    curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.5.9 1.5.7
else
    echo "✅ Fabric tools found."
fi

export PATH=${PWD}/bin:$PATH
export FABRIC_CFG_PATH=${PWD}/config/

# --------------------------------------------------------
# 2. تشغيل الشبكة
# --------------------------------------------------------
echo -e "${GREEN}🌐 Step 2: Starting Fabric Network...${NC}"
cd test-network
./network.sh down
./network.sh up createChannel -c mychannel -ca
cd ..

# --------------------------------------------------------
# 3. نشر العقد الذكي
# --------------------------------------------------------
echo -e "${GREEN}📜 Step 3: Deploying Smart Contract (Go)...${NC}"
cd test-network
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
cd ..

# --------------------------------------------------------
# 4. إعداد وتشغيل Caliper (الجزء الذكي)
# --------------------------------------------------------
echo -e "${GREEN}⚡ Step 4: Configuring & Running Caliper...${NC}"
cd caliper-workspace

# أ) تثبيت المكتبات إذا لم تكن موجودة
if [ ! -d "node_modules" ]; then
    echo "📦 Installing Caliper dependencies..."
    npm install
    npx caliper bind --caliper-bind-sut fabric:2.2
fi

# ب) البحث عن المفتاح الخاص (Private Key) أوتوماتيكياً
echo "🔑 Detecting Private Key..."
KEY_DIR="../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore"
PVT_KEY=$(ls $KEY_DIR/*_sk)
echo "✅ Found Key: $PVT_KEY"

# ج) إنشاء ملف إعدادات الشبكة بالمسار الصحيح
echo "⚙️ Generating network config..."
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
        - name: 'User1'
          clientPrivateKey:
            path: '$PVT_KEY'
          clientSignedCert:
            path: '../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/cert.pem'
    connectionProfile:
      path: '../test-network/organizations/peerOrganizations/org1.example.com/connection-org1.yaml'
      discover: true
EOF
echo -e "nawal2"
# د) تشغيل الاختبار
echo "🔥 Running Benchmarks..."
npx caliper launch manager \
    --caliper-workspace . \
    --caliper-networkconfig networks/networkConfig.yaml \
    --caliper-benchconfig benchmarks/benchConfig.yaml \
    --caliper-flow-only-test

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}🎉 Project Finished Successfully!${NC}"
echo -e "${GREEN}📄 Report: caliper-workspace/report.html${NC}"
