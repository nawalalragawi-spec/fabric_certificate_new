'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class IssueCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;

        // 1. تصحيح: استخدام Backticks (`) لدمج المتغيرات داخل النص
        const certID = `CERT_${this.workerIndex}_${this.txIndex}`;
        const studentName = `Student_Name_${this.workerIndex}_${this.txIndex}`;
        const degree = 'Bachelor of Science in Information Technology';
        const issuer = 'Universiti Utara Malaysia (UUM)';
        const issueDate = new Date().toISOString();

        // 2. إعداد الطلب (متوافق تماماً مع وسائط العقد الذكي Go)
        const request = {
            contractId: 'basic',
            contractFunction: 'IssueCertificate',
            contractArguments: [
                certID,
                studentName,  // سيتم تشفيره في الـ Chaincode
                degree,
                issueDate,
                issuer
            ],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(request);
    }
}

// 3. إضافة دالة التصدير (ضرورية جداً لكي يتعرف Caliper على الملف)
function createWorkloadModule() {
    return new IssueCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
