'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const crypto = require('crypto'); // مكتبة التشفير المدمجة في Node.js

class IssueCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        
        // 1. إنشاء بيانات فريدة
        const certID = `CERT_${this.workerIndex}_${this.txIndex}`;
        const studentName = `Student_${this.workerIndex}_${this.txIndex}`;
        
        // 2. محاكاة التشفير (EdDSA/AES) كما في دراسة عمر سعد
        // نقوم بإنشاء Hash معقد يمثل التوقيع الرقمي للشهادة
        const signature = crypto.createHmac('sha256', 'secret-key')
                               .update(certID + studentName)
                               .digest('hex');

        // 3. زيادة حجم البيانات (Payload) لمحاكاة شهادة كاملة
        const degree = 'Bachelor of Science in Information Technology';
        const issuer = 'Universiti Utara Malaysia (UUM)'; // محاكاة لجهة الإصدار في الدراسة
        const certHash = signature; // استخدام التوقيع كـ Hash للشهادة
        const issueDate = new Date().toISOString();

        const request = {
            contractId: 'basic', 
            contractFunction: 'IssueCertificate', 
            contractArguments: [
                certID,       
                studentName,  
                degree,       
                issuer,       
                certHash,     
                issueDate     
            ],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new IssueCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule; 
