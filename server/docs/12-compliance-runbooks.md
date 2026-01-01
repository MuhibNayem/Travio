# Nationwide Ticketing System - Compliance and Operational Runbooks

## 1) PCI-DSS Runbook (Payments)

Objectives:
- Limit card data exposure to the Payment Service.
- Enforce tokenization and avoid PAN storage.

Controls:
- Network segmentation around Payment Service.
- Mandatory encryption in transit and at rest.
- Short-lived access tokens and strict IAM.

Operational Steps:
1) Quarterly ASV scans and remediation.
2) Monthly access review for payment systems.
3) Rotate secrets on schedule (30-90 days).
4) Verify logs for payment events are immutable.

## 2) GDPR Runbook (Privacy)

Objectives:
- Support data access and deletion requests.
- Minimize personal data retained.

Operational Steps:
1) Verify identity of requester.
2) Collect data from Identity, Orders, Notifications, and Logs.
3) Perform deletion/anonymization in downstream systems.
4) Log completion with audit trail and retention policy.

## 3) Data Retention Policy

- Orders and payments: 7 years (finance/legal).
- User profile data: active + 24 months after inactivity.
- Logs and traces: 90 days (hot), 1 year (cold).

Operational Steps:
1) Identify partitions older than retention window.
2) Export for legal hold if required.
3) Delete or archive partitions.

## 4) Incident Response Runbook

Severity levels:
- SEV-1: checkout outage or oversell.
- SEV-2: partial service degradation.
- SEV-3: minor functional issues.

Steps:
1) Declare incident in on-call channel.
2) Assign incident commander and comms lead.
3) Triage systems: API gateway, inventory, payments.
4) Mitigate using feature flags or queue throttling.
5) Postmortem within 48 hours.

## 5) Access Control and Audit

- All access logged with user, action, and reason.
- Production access requires MFA and approval.
- Privileged access reviewed monthly.

## 6) Vendor Compliance Runbook

Objectives:
- Ensure vendor onboarding complies with KYC/AML obligations.
- Maintain contract acceptance and payout verification records.

Operational Steps:
1) Verify KYC provider webhook signatures and status.
2) Review vendor contract acceptance and revenue share terms.
3) Confirm payout account verification status.
4) Record audit event and retain proof bundles.
