'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class IssueCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        
        // إنشاء بيانات تجريبية فريدة لكل عملية إصدار
        const certID = `CERT_${this.workerIndex}_${this.txIndex}`;
        const studentName = `Student_${this.workerIndex}_${this.txIndex}`;
        const degree = 'Bachelor of Computer Science';
        const issuer = 'Digital University';
        // محاكاة بصمة رقمية (Hash) فريدة
        const certHash = Buffer.from(certID + studentName).toString('hex'); 
        const issueDate = '2026-01-02';

        const request = {
            contractId: 'basic', // تأكد أن هذا الاسم يطابق اسم العقد في ملف التكوين
            contractFunction: 'IssueCertificate', // اسم الدالة الجديدة في Go
            contractArguments: [
                certID,       // ID
                studentName,  // studentName
                degree,       // degree
                issuer,       // issuer
                certHash,     // certHash
                issueDate     // issueDate
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
