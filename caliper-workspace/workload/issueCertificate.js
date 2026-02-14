'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const crypto = require('crypto');

// مفتاح التشفير (يجب أن يطابق الموجود في العقد الذكي لضمان نجاح فك التشفير لاحقاً)
const AES_KEY = Buffer.from('12345678901234567890123456789012');

function encrypt(text) {
    const iv = crypto.randomBytes(12);
    const cipher = crypto.createCipheriv('aes-256-gcm', AES_KEY, iv);
    let encrypted = cipher.update(text, 'utf8', 'hex');
    encrypted += cipher.final('hex');
    const authTag = cipher.getAuthTag().toString('hex');
    return `${iv.toString('hex')}:${encrypted}:${authTag}`;
}

class IssueCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        const certID = `cert_${this.workerIndex}_${this.txIndex}`;
        
        // البيانات المراد حمايتها
        const originalHash = "hash_data_" + this.txIndex;
        const encryptedHash = encrypt(originalHash);

        /**
         * ملاحظة هامة للمطابقة مع ورقة عمر سعد:
         * 1. يجب أن يكون المستدعي (Invoker) يملك دور "admin" في ملف networkConfig.
         * 2. الـ Owner هنا يفضل أن يكون هوية الطالب الافتراضية في الشبكة.
         */
        const studentID = `x509::/C=US/ST=North Carolina/L=Durham/O=org1.example.com/CN=User1::/C=US/ST=North Carolina/L=Durham/O=org1.example.com/CN=ca.org1.example.com`;

        const request = {
            contractId: 'basic', // تأكد من مطابقة اسم العقد المنصب
            contractFunction: 'IssueCertificate',
            contractArguments: [
                certID,
                studentID,      // المالك (Owner) - ضروري لعملية الـ Unlock لاحقاً
                'University_A', // المصدر (Issuer)
                '2025-05-20',   // تاريخ الإصدار
                encryptedHash   // الهاش المشفر بـ AES256
            ],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() { return new IssueCertificateWorkload(); }
module.exports.createWorkloadModule = createWorkloadModule;
