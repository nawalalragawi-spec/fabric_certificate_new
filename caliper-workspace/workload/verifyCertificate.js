'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class VerifyCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        // استخدام نفس نمط المعرف المستخدم في عملية الإصدار
        const certID = `cert_${this.workerIndex}_${this.txIndex}`;

        const request = {
            contractId: 'basic',
            // التعديل 1: تغيير اسم الدالة إلى VerifyCertificate الموجودة في Go
            contractFunction: 'VerifyCertificate', 
            // التعديل 2: الدالة تتوقع وسيطين (ID و Hash) حسب العقد الذكي
            contractArguments: [
                certID, 
                'dummy_hash_for_now' // يجب أن يطابق الهاش المستخدم عند الإصدار
            ],
            readOnly: true
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new VerifyCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;