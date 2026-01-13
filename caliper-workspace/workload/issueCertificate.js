'use strict';
const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
class IssueCertificateWorkload extends WorkloadModuleBase {
constructor() {
super();
this.txIndex = 0;
}
async submitTransaction() {
this.txIndex++;

// 1. إنشاء بيانات فريدة لكل معاملة
const certID = `CERT_${this.workerIndex}_${this.txIndex}`;
const studentName = `Student_Name_${this.workerIndex}_${this.txIndex}`;
const degree = 'Bachelor of Science in Information Technology';
const issuer = 'Universiti Utara Malaysia (UUM)'; 
const issueDate = new Date().toISOString();

// 2. إعداد الطلب ليتوافق تماماً مع وسائط العقد الذكي الجديد
// الترتيب: [id, studentName, degree, issueDate, issuer]
const request = {
    contractId: 'basic', 
    contractFunction: 'IssueCertificate', 
    contractArguments: [
        certID,       
        studentName,  // العقد سيقوم بتشفير هذا الحقل بـ AES-256 داخلياً
        degree,       
        issueDate,    
        issuer        
    ],
    readOnly: false
};

await this.sutAdapter.sendRequests(request); 
