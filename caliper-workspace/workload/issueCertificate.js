// IssueCertificate: التحقق من دور المسؤول وإصدار الشهادة بحالة مقفلة
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, owner string, issuer string, issueDate string, encryptedHash string) error {
    
    // 1. التحقق من صلاحية "الآدمن" (ABAC) لضمان أن الجامعة فقط من تصدر الشهادات
    err := ctx.GetClientIdentity().AssertAttributeValue("role", "admin")
    if err != nil {
        return fmt.Errorf("غير مصرح: فقط المستخدمين بصلاحية admin يمكنهم إصدار الشهادات")
    }

    // 2. التحقق من عدم وجود الشهادة مسبقاً
    exists, err := s.CertificateExists(ctx, id)
    if err != nil { return err }
    if exists { return fmt.Errorf("الشهادة %s موجودة بالفعل", id) }

    // 3. بناء هيكل الشهادة (ضبط IsLocked على true افتراضياً كخيار خصوصية في ورقة عمر سعد)
    certificate := Certificate{
        ID:        id,
        Owner:     owner,         // هذا هو المعرف الذي سيستخدمه Caliper في الـ Unlock
        Issuer:    issuer,
        IssueDate: issueDate,
        CertHash:  encryptedHash, // الهاش المشفر بـ AES
        IsLocked:  true,          // الحالة الافتراضية مقفلة
    }

    certBytes, err := json.Marshal(certificate)
    if err != nil { return err }

    // 4. الحفظ في سجل البلوكشين
    return ctx.GetStub().PutState(id, certBytes)
}
