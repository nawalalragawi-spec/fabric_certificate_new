'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class IssueCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        
        // إنشاء معرف فريد لكل معاملة
        const certID = `cert_${this.workerIndex}_${this.txIndex}`;
        
        const request = {
            // تأكد أن contractId يطابق الاسم في networkConfig.yaml (غالباً basic)
            contractId: 'basic', 
            // 1. تصحيح اسم الدالة لتطابق كود الـ Go
            contractFunction: 'IssueCertificate', 
            // 2. تصحيح الوسائط لتطابق ترتيب وأنواع كود الـ Go
            contractArguments: [
                certID,                   // id
                'Student_' + this.txIndex, // owner
                'University_A',           // issuer
                '2025-05-20',             // issueDate
                'dummy_hash_for_now'      // certHash (سنضيف التشفير الفعلي لاحقاً)
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