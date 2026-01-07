'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const crypto = require('crypto');

class IssueCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }

    async initializeWorkloadModule(workerIndex, totalWorkers, numberProtocols, workloadContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, numberProtocols, workloadContext);
    }

    async submitTransaction() {
        const certId = `Cert_${this.workerIndex}_${Date.now()}_${Math.floor(Math.random() * 1000)}`;
        const owner = 'Student_' + Math.floor(Math.random() * 1000);
        const issuer = 'University_Authority';
        const issueDate = new Date().toISOString();
        const randomContent = crypto.randomBytes(20).toString('hex');
        const certHash = crypto.createHash('sha256').update(randomContent).digest('hex');

        // هذا هو الجزء الذي كان يسبب الخطأ في الصورة
        const request = {
            contractId: 'basic', // تأكد أن هذا هو اسم الـ Chaincode المسجل في الشبكة
            contractFunction: 'CreateCertificate', // تأكد من مطابقة اسم الدالة في Go
            contractArguments: [certId, owner, issuer, issueDate, certHash],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new IssueCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
