'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class VerifyCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        
        // يجب أن يتطابق المعرف مع ما تم إصداره لضمان وجود السجل
        const certID = `CERT_${this.workerIndex}_${this.txIndex}`;

        const request = {
            contractId: 'basic',
            // استدعاء دالة القراءة التي تتضمن فك التشفير (AES Decrypt)
            contractFunction: 'ReadCertificate', 
            contractArguments: [certID],
            readOnly: true // عملية قراءة فقط (Query)
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new VerifyCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
