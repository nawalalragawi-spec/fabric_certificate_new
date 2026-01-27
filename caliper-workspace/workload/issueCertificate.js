// ستحتاج لإضافة هذه الدوال المساعدة في ملف Go الخاص بك
import (
    "crypto/aes"
    "crypto/cipher"
    "encoding/hex"
    "strings"
)

const AES_KEY = "12345678901234567890123456789012" // يجب أن يطابق المفتاح في JS

func decrypt(encryptedData string) (string, error) {
    parts := strings.Split(encryptedData, ":")
    iv, _ := hex.DecodeString(parts[0])
    ciphertext, _ := hex.DecodeString(parts[1])
    authTag, _ := hex.DecodeString(parts[2])

    block, _ := aes.NewCipher([]byte(AES_KEY))
    aesgcm, _ := cipher.NewGCM(block)

    // دمج التشفير مع الـ tag لفكها في Go
    fullCiphertext := append(ciphertext, authTag...)
    plaintext, err := aesgcm.Open(nil, iv, fullCiphertext, nil)
    if err != nil {
        return "", err
    }
    return string(plaintext), nil
}

// ثم نعدل دالة التحقق لتستخدم فك التشفير
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string, providedHash string) (bool, error) {
    certificate, err := s.ReadCertificate(ctx, id)
    if err != nil { return false, err }

    // فك تشفير الهاش المخزن قبل المقارنة
    decryptedStoredHash, err := decrypt(certificate.CertHash)
    if err != nil { return false, fmt.Errorf("failed to decrypt: %v", err) }

    return decryptedStoredHash == providedHash, nil
}
