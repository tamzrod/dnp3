# Glossary

## Purpose

This glossary establishes canonical terminology for KDSE. Each term has exactly one definition within the KDSE context. When the same term has different meanings in different contexts, this glossary establishes the authoritative meaning.

## A

### Adoption

The process by which a team establishes and practices KDSE. Adoption involves understanding principles, establishing governance, and applying KDSE to real decisions.

### Alignment

The state of conformance between artifacts. Implementation aligns with Architecture when implementation follows architectural direction. Architecture aligns with Knowledge when architecture derives from and satisfies knowledge requirements.

### Architecture Decision Record (ADR)

An artifact documenting a significant architectural decision. An ADR captures what was decided, why it was decided, what alternatives were considered, and what consequences the decision has.

### Artifact

A tangible or intangible object created, maintained, or used during the engineering process.

In KDSE, artifacts are typed (Knowledge, Architecture, Implementation, Verification, Governance) with defined purposes, stewards, lifetimes, and authority levels.

### Artifact Lifecycle

The progression of artifacts through defined states. Artifact lifecycle manages artifact quality, review requirements, and authority levels through states such as Draft, Reviewed, Approved, and Superseded.

### Artifact State

The current position of an artifact within its lifecycle. Artifact state communicates readiness and authority level. States include Proposed, Draft, Reviewed, Approved, Reference, Canonical, Superseded, Deprecated, and Archived.

### Authority

The legitimate power to make decisions and the legitimacy to require compliance.

In KDSE, authority flows downward through the artifact hierarchy. Higher-authority artifacts constrain and authorize lower-authority artifacts.

### Authority Hierarchy

The structure of authority in KDSE, from highest (Knowledge) to lowest (Implementation), with Verification reporting on alignment.

### Authority Resolution

The process of resolving conflicts between artifacts while maintaining the authority hierarchy. Authority resolution may involve correcting the lower artifact, clarifying the higher artifact, or adding intermediate knowledge.

## B

### Body of Knowledge (BoK)

The complete set of concepts, terms, and activities that constitute a discipline.

The KDSE Body of Knowledge encompasses all artifacts, principles, processes, and practices that define the discipline.

## C

### Case Study

An documented account of KDSE application that provides evidence of methodology effectiveness. Case studies validate KDSE; they do not define it.

### Change Management

The process by which changes to artifacts are proposed, evaluated, approved, implemented, and documented.

In KDSE, changes must respect the authority hierarchy. Changes to lower-layer artifacts do not automatically propagate, but changes to higher-layer artifacts may require corresponding lower-layer changes.

### Context

The circumstances, environment, and conditions within which a system operates.

In KDSE, context is captured as knowledge and informs architectural decisions.

## D

### Derivation

The process by which higher-authority artifacts produce lower-authority artifacts through analysis, decision-making, and documentation. Derivation transforms validated understanding into structural decisions.

In KDSE, derivation is mandatory: Architecture derives from Knowledge; Implementation derives from Architecture.

### Discipline

A field of study or professional practice governed by established principles and practices.

KDSE is a discipline within software engineering.

## E

### Evolution

The process by which artifacts change over time while maintaining traceability and authority alignment.

Evolution is the fifth stage of the KDSE lifecycle and is continuous throughout the system lifecycle.

## G

### Governance

The system of rules, practices, and processes by which authority is exercised and decisions are made.

In KDSE, Governance artifacts establish stewardship, authority delegation, and process compliance requirements.

## I

### Implementation

The physical realization of a system through code, configuration, and related artifacts.

Implementation derives authority from Architecture and must not contradict Architecture or Knowledge.

### Intent

The purpose or goal behind a decision, requirement, or artifact.

In KDSE, understanding intent is essential for maintaining alignment when deriving artifacts.

## K

### Knowledge

The authoritative understanding about a problem domain, including requirements, constraints, context, and assumptions.

Knowledge is the highest-authority artifact type in KDSE. All other artifacts derive authority from Knowledge.

### Knowledge Artifact

An artifact of type Knowledge, capturing authoritative understanding.

Knowledge artifacts are created during the Knowledge stage and persist throughout the system lifecycle.

### Knowledge Steward

The individual or role responsible for ensuring knowledge artifacts are created, maintained, validated, and evolved. The Knowledge Steward is responsible for the quality and authority of knowledge artifacts.

## M

### Methodology Maturity Model

A framework for assessing the development state of a methodology. KDSE defines six maturity levels: Concept, Defined, Structured, Usable, Validated, and Proven.

## N

### Non-Functional Requirement

A requirement that specifies criteria for system operation rather than specific behavior.

Examples include performance, security, reliability, and scalability requirements. Non-functional requirements are captured as Knowledge artifacts.

## O

### Obsolescence

The state of knowledge when it no longer accurately represents the problem domain. Obsolete knowledge loses authority but is maintained for traceability.

## P

### Pattern

A reusable solution to a recurring problem within a specific context.

Patterns may inform architecture and implementation but are not authoritative in KDSE. Authority derives from Knowledge artifacts.

### Practice

An activity or technique performed as part of engineering work.

Practices in KDSE are derived from principles and may vary by context. Practices are not authoritative; principles are.

### Principle

A fundamental truth or proposition that guides decision-making and action.

KDSE defines core principles that are timeless and context-independent.

### Problem Domain

The area of expertise or application to which a system addresses.

Understanding the problem domain is captured as Knowledge.

### Provenance

The origin and history of knowledge. Provenance includes the sources of knowledge, the validation process, and the changes the knowledge has undergone.

## R

### Requirement

A statement of need or constraint that a system must satisfy.

Requirements are captured as Knowledge artifacts in KDSE.

### Repository

A storage location for artifacts, typically version-controlled.

In KDSE, a repository may contain artifacts of any type. The repository maintains traceability relationships between artifacts.

### Resolution

The process of addressing a contradiction or gap between artifacts.

Resolution may involve clarification at a higher layer or correction at a lower layer.

## S

### Steward

An individual or role responsible for artifact quality, maintenance, and evolution. Stewardship replaces ownership-oriented thinking in KDSE.

### Stewardship

The responsibility for artifacts that emphasizes care and maintenance over possession. Knowledge should be stewarded, not owned.

### Stewardship Transfer

The process of handing over artifact responsibility from one steward to another. Stewardship transfer requires documentation, knowledge transfer, acknowledgment, and stakeholder notification.

### Structured Knowledge

Engineering knowledge that has been validated, documented with dependencies and provenance, assigned stewardship, and defined with explicit lifetime boundaries. Only structured knowledge can carry authority in KDSE.

## T

### Traceability

The ability to follow relationships between artifacts. Traceability enables understanding of how artifacts relate to each other and why they exist.

In KDSE, every artifact below Knowledge must be traceable to authorized Knowledge artifacts. Traceability enables impact analysis and compliance verification.

### Traceability Path

The sequence of artifacts connecting two related artifacts.

For example, an Implementation artifact traces to an Architecture artifact through an explicit relationship. The Architecture artifact traces to a Knowledge artifact.

### Traceability Link

An explicit relationship between two artifacts. Traceability links form the chains that traceability follows.

### Traceability Verification

The process of confirming that traceability links are complete, correct, consistent, and current.

## V

### Verification

The process of confirming that implementation aligns with architecture and that architecture aligns with knowledge.

Verification artifacts document the results of verification activities. Verification derives authority from Knowledge artifacts.

### Verification Artifact

An artifact of type Verification, documenting verification activities and results.

Verification artifacts include verification plans, test cases, test results, and reports.

### Verification Domain

The knowledge domain encompassing verification principles, processes, and criteria. The verification domain includes verification goals, evidence requirements, traceability, criteria, authority, and lifecycle.

### Verification Criteria

The standards by which verification is evaluated. Verification criteria include completeness, correctness, consistency, reproducibility, and independence.

### Verification Steward

The individual or role responsible for ensuring verification activities are properly planned, executed, and documented.

### Verification State

The classification of verification evidence status. KDSE defines four verification states: Verified, Verified with Failures, Not Verified, and Not Assessed. The verification state determines scoring caps and risk levels.

### Test Asset

An artifact related to testing that defines, plans, or documents tests but does not prove tests were executed. Test assets include verification plans, test cases, and test documentation. Test assets are NOT verification evidence.

### Test Execution Evidence

Records that prove tests were actually executed. Test execution evidence includes test results, execution logs, CI/CD build logs, and non-conformance reports. Test execution evidence is REQUIRED for verification claims.

### Verified

A verification state indicating that tests were executed and passing results were documented. Verified status requires both test execution records and test results showing passing outcomes.

### Verified with Failures

A verification state indicating that tests were executed but failures were documented. Verified with Failures status requires both test execution records and test results showing failures.

### Not Verified

A verification state indicating that test assets exist (test plans, test cases, test documentation) but no execution evidence exists (no test results, no execution logs). Not Verified status caps scoring at 4/10 and indicates HIGH risk.

### Not Assessed

A verification state indicating that no verification artifacts exist. Not Assessed status caps scoring at 2/10 and indicates MAXIMUM risk.

## W

### Working Memory

The temporary state used during reasoning and derivation.

Working memory in KDSE refers to intermediate artifacts and reasoning processes used during derivation. Working memory is distinct from persistent artifacts.

---

## Version

- **Document Version**: 1.1
- **Effective Date**: 2026-07-10
- **Change Note**: Added verification evidence terminology (Verification State, Test Asset, Test Execution Evidence, Verified, Verified with Failures, Not Verified, Not Assessed) to address KDSE-DEFECT-001
