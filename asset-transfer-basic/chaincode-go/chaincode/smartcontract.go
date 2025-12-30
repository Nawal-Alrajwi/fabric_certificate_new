package chaincode
import (
"encoding/hex"
"encoding/json"
"fmt"

"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
"golang.org/x/crypto/sha3" // استخدام SHA-3 المتطورة لمنافسة SHA-256
)
type SmartContract struct {
contractapi.Contract
}
// الهيكل المطور للشهادة (يتضمن البصمة الرقمية)
type Certificate struct {
ID          string json:"ID"
StudentName string json:"StudentName"
Major       string json:"Major"
University  string json:"University"
IssueDate   string json:"IssueDate"
Grade       string json:"Grade"
IssuerID    string json:"IssuerID"
CertHash    string json:"CertHash" // الحقل المضاف لتقليل زمن التحقق
}
// دالة داخلية لتوليد بصمة SHA-3 (Keccak-256)
func calculateSHA3Hash(data string) string {
hash := sha3.New256()
hash.Write([]byte(data))
return hex.EncodeToString(hash.Sum(nil))
}
// 1. إضافة شهادة جديدة مع توليد بصمة SHA-3
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, studentName string, major string, university string, issueDate string, grade string, issuerID string) error {
exists, err := s.CertificateExists(ctx, id)
if err != nil {
return err
}
if exists {
return fmt.Errorf("الشهادة ذات الرقم %s موجودة مسبقاً", id)
}

// دمج البيانات لتوليد البصمة الفريدة
combinedData := fmt.Sprintf("%s%s%s%s", id, studentName, university, issueDate)
certHash := calculateSHA3Hash(combinedData)

cert := Certificate{
	ID:          id,
	StudentName: studentName,
	Major:       major,
	University:  university,
	IssueDate:   issueDate,
	Grade:       grade,
	IssuerID:    issuerID,
	CertHash:    certHash,
}

certJSON, err := json.Marshal(cert)
if err != nil {
	return err
}

return ctx.GetStub().PutState(id, certJSON)
}
// 2. دالة الحذف (Revoke) - المعالجة الأساسية لنقاط الفشل في تقريرك
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
exists, err := s.CertificateExists(ctx, id)
if err != nil {
return err
}
if !exists {
return fmt.Errorf("الشهادة %s غير موجودة ولا يمكن حذفها", id)
}

// حذف الحالة من السجل (Ledger)
return ctx.GetStub().DelState(id)
}
// 3. التحقق السريع من الشهادة (اثبات تفوق زمن الاستجابة)
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string, providedData string) (bool, error) {
certJSON, err := ctx.GetStub().GetState(id)
if err != nil {
return false, fmt.Errorf("فشل في قراءة البيانات: %v", err)
}
if certJSON == nil {
return false, fmt.Errorf("الشهادة %s غير موجودة", id)
}

var cert Certificate
err = json.Unmarshal(certJSON, &cert)
if err != nil {
	return false, err
}

// مقارنة البصمة المدخلة مع البصمة المخزنة SHA-3
currentHash := calculateSHA3Hash(providedData)
return cert.CertHash == currentHash, nil
}
// 4. استعلام مطور يدعم التقسيم (Pagination) ليتوافق مع ملف Caliper JS
func (s *SmartContract) GetAllAssetsWithPagination(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (*PaginatedQueryResult, error) {
    // جلب البيانات على دفعات لتقليل زمن الاستجابة ومنع Timeout
    resultsIterator, responseMetadata, err := ctx.GetStub().GetStateByRangeWithPagination("", "", pageSize, bookmark)
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var certs []*Certificate
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }
        var cert Certificate
        err = json.Unmarshal(queryResponse.Value, &cert)
        if err != nil {
            return nil, err
        }
        certs = append(certs, &cert)
    }

    return &PaginatedQueryResult{
        Records:             certs,
        FetchedRecordsCount: responseMetadata.FetchedRecordsCount,
        Bookmark:            responseMetadata.Bookmark,
    }, nil
}

// تعريف هيكل البيانات المطلوب للرد المقسم
type PaginatedQueryResult struct {
    Records             []*Certificate `json:"records"`
    FetchedRecordsCount int32          `json:"fetchedRecordsCount"`
    Bookmark            string         `json:"bookmark"`
}
// 5. دالة التحقق من الوجود
func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
certJSON, err := ctx.GetStub().GetState(id)
if err != nil {
return false, err
}
return certJSON != nil, nil
} 
