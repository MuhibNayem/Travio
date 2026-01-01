# Nationwide Ticketing System - RACI and Staffing Plan

This plan maps roles to roadmap sprints and clarifies ownership.

## 1) Roles

- Platform Lead (PL): infrastructure, Kubernetes, service mesh.
- Backend Lead (BL): service design and Go standards.
- Data Lead (DL): ScyllaDB, PostgreSQL, Kafka pipelines.
- Security Lead (SL): PCI, audits, IAM.
- QA Lead (QL): test automation and performance testing.
- Product Lead (PRL): scope, requirements, prioritization.
- SRE Lead (SRL): SLOs, observability, incident response.

## 2) Staffing Assumptions

- 2 squads for Core Services (Identity, Catalog, Inventory, Order).
- 1 squad for Data/Streaming (Kafka, Search, Reporting).
- 1 squad for Risk/Compliance (Fraud, Audit, Payments).
- 1 platform squad (K8s, CI/CD, Observability).

## 3) RACI Matrix by Sprint

Legend: R = Responsible, A = Accountable, C = Consulted, I = Informed

Sprint 0:
- Platform setup: PL(A), SRL(R), BL(C), DL(C), SL(C), QL(I), PRL(I)

Sprint 1:
- Identity/Catalog: BL(A), Core Squad(R), SL(C), QL(C), PRL(C), SRL(I)

Sprint 2:
- Inventory: BL(A), Core Squad(R), DL(C), SRL(C), QL(C), PRL(I)

Sprint 3:
- Orders/Checkout: BL(A), Core Squad(R), DL(C), QL(C), SL(C), PRL(I)

Sprint 4:
- Payments/Fraud: SL(A), Risk Squad(R), BL(C), DL(C), QL(C), PRL(I)

Sprint 5:
- Fulfillment/Notifications: BL(A), Core Squad(R), QL(C), PRL(C)

Sprint 6:
- Search: DL(A), Data Squad(R), BL(C), QL(C)

Sprint 7:
- Queue/Rate Limiting: PL(A), Platform Squad(R), SRL(C), BL(C)

Sprint 8:
- Multi-Region: SRL(A), Platform Squad(R), DL(C), BL(C), SL(C)

Sprint 9:
- Pricing/Promotions: PRL(A), Core Squad(R), BL(C), QL(C)

Sprint 10:
- Reporting/Analytics: DL(A), Data Squad(R), PRL(C), QL(C)

Sprint 11:
- Compliance/Hardening: SL(A), SRL(R), QL(C), BL(C), PRL(I)

## 4) Hiring Guidance

Short-term (0-3 months):
- 1 Platform Engineer, 2 Backend Engineers, 1 Data Engineer, 1 QA.

Mid-term (3-6 months):
- 1 SRE, 1 Security Engineer, 1 Fraud Specialist.

Long-term (6-12 months):
- 2 Backend Engineers, 1 Data Scientist, 1 Product Analyst.

