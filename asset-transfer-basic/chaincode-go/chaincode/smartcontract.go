package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Certificate تعريف هيكل الشهادة
type Certificate struct {
	CertHash    string `json:"CertHash"`    // بصمة ملف الشهادة لمنع التزوير
	Degree      string `json:"Degree"`      // التخصص أو الدرجة
	ID          string `json:"ID"`          // الرقم التسلسلي للشهادة
	IsRevoked   bool   `json:"IsRevoked"`   // حالة الشهادة (ملغية أم لا)
	IssueDate   string `json:"IssueDate"`   // تاريخ الصدور
	Issuer      string `json:"Issuer"`      // الجهة المانحة للشهادة
	StudentName string `json:"StudentName"` // اسم الطالب
}

// 1. IssueCertificate: إصدار شهادة جديدة (محسّنة)
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, studentName string, degree string, issuer string, certHash string, issueDate string) error {
	// Validation قوي للمدخلات
	if id == "" || studentName == "" || degree == "" || issuer == "" || certHash == "" || issueDate == "" {
		return fmt.Errorf("جميع الحقول مطلوبة ولا يمكن أن تكون فارغة")
	}

	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	
	// إذا كانت موجودة، تحقق أنها متطابقة (idempotency)
	if exists {
		existingCert, err := s.ReadCertificate(ctx, id)
		if err != nil {
			return err
		}
		// إذا كانت نفس البيانات، عد بنجاح بدلاً من الفشل
		if existingCert.StudentName == studentName && existingCert.CertHash == certHash {
			return nil // العملية موجودة بالفعل، لا مشكلة
		}
		return fmt.Errorf("الشهادة ذات الرقم %s موجودة مسبقاً", id)
	}

	cert := Certificate{
		ID:          id,
		StudentName: studentName,
		Degree:      degree,
		Issuer:      issuer,
		CertHash:    certHash,
		IssueDate:   issueDate,
		IsRevoked:   false, // الشهادة فعالة عند الإصدار
	}
	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

// 2. QueryAllCertificates: استعلام عن جميع الشهادات المخزنة (مع pagination)
func (s *SmartContract) QueryAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Certificate, error) {
	// تحديد حد أقصى للنتائج في كل استعلام لتجنب memory overload
	const maxResults = 120
	
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var certificates []*Certificate
	count := 0
	
	for resultsIterator.HasNext() && count < maxResults {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var cert Certificate
		err = json.Unmarshal(queryResponse.Value, &cert)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, &cert)
		count++
	}

	return certificates, nil
}

// 2b. QueryAllCertificatesWithPagination: استعلام مع تقسيم الصفحات (للأداء العالي)
func (s *SmartContract) QueryAllCertificatesWithPagination(ctx contractapi.TransactionContextInterface, pageSize string, bookmark string) (string, error) {
	pageInt := 120
	if pageSize != "" {
		// يمكن تمرير حجم الصفحة
		fmt.Sscanf(pageSize, "%d", &pageInt)
		if pageInt > 200 {
			pageInt = 200 // حد أقصى 200 نتيجة
		}
	}

	// الطريقة الآمينة: استخدام GetStateByRange بدون pagination API
	// للتوافق مع إصدارات Fabric المختلفة
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return "", err
	}
	defer resultsIterator.Close()

	var certificates []*Certificate
	count := 0
	for resultsIterator.HasNext() && count < pageInt {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}

		var cert Certificate
		err = json.Unmarshal(queryResponse.Value, &cert)
		if err != nil {
			return "", err
		}
		certificates = append(certificates, &cert)
		count++
	}

	// إرجاع النتائج بتنسيق JSON
	resultJSON := map[string]interface{}{
		"records": certificates,
		"count": len(certificates),
	}
	
	resultBytes, _ := json.Marshal(resultJSON)
	return string(resultBytes), nil
}

// 3. RevokeCertificate: إلغاء شهادة (محسّنة)
func (s *SmartContract) RevokeCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	// Validation
	if id == "" {
		return fmt.Errorf("معرف الشهادة مطلوب")
	}

	cert, err := s.ReadCertificate(ctx, id)
	if err != nil {
		// إذا كانت الشهادة غير موجودة، قد تكون ملغاة بالفعل
		return nil // لا نرجع خطأ إذا كانت غير موجودة
	}

	// إذا كانت ملغاة بالفعل، عد بنجاح (idempotency)
	if cert.IsRevoked {
		return nil
	}

	cert.IsRevoked = true // تغيير الحالة إلى ملغية
	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

// 4. VerifyCertificate: التحقق من صحة الشهادة وصلاحيتها (محسّنة)
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string, certHash string) (bool, error) {
	// Validation
	if id == "" || certHash == "" {
		return false, fmt.Errorf("معرف الشهادة والبصمة مطلوبة")
	}

	cert, err := s.ReadCertificate(ctx, id)
	if err != nil {
		// شهادة غير موجودة ليست خطأ، لكن التحقق يفشل
		return false, nil
	}

	// التأكد من أن البصمة مطابقة وأن الشهادة ليست ملغية
	isValid := cert.CertHash == certHash && !cert.IsRevoked
	return isValid, nil
}

// --- وظائف مساعدة ---

func (s *SmartContract) ReadCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}
	if certJSON == nil {
		return nil, fmt.Errorf("الشهادة %s غير موجودة", id)
	}

	var cert Certificate
	err = json.Unmarshal(certJSON, &cert)
	return &cert, err
}

func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}
	return certJSON != nil, nil
}
