'use strict';
const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const crypto = require('crypto');

// مفتاح التشفير (يجب أن يكون 32 بايت لـ AES-256)
const AES_KEY = Buffer.from('12345678901234567890123456789012'); // مثال ثابت للتبسيط حالياً

function encrypt(text) {
    const iv = crypto.randomBytes(12); // GCM يتطلب 12 بايت IV
    const cipher = crypto.createCipheriv('aes-256-gcm', AES_KEY, iv);
    let encrypted = cipher.update(text, 'utf8', 'hex');
    encrypted += cipher.final('hex');
    const authTag = cipher.getAuthTag().toString('hex');
    // ندمج (iv + encrypted + authTag) في سلسلة واحدة لإرسالها
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
        
        // البيانات المراد حمايتها (الهاش الأصلي)
        const originalHash = "sha256_of_real_certificate_data";
        // تشفير الهاش قبل الإرسال
        const encryptedHash = encrypt(originalHash);

        const request = {
            contractId: 'basic',
            contractFunction: 'IssueCertificate',
            contractArguments: [
                certID,
                'Student_' + this.txIndex,
                'University_A',
                '2025-05-20',
                encryptedHash // نرسل الهاش مشفراً
            ],
            readOnly: false
        };
        await this.sutAdapter.sendRequests(request);
    }
}
function createWorkloadModule() { return new IssueCertificateWorkload(); }
module.exports.createWorkloadModule = createWorkloadModule;
