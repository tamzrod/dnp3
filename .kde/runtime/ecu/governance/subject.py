"""
Subject Classifier for Laboratory Governance
"""
from enum import Enum
from typing import Dict, List
from dataclasses import dataclass

class Subject(Enum):
    KDE = "kde"
    PROJECT = "project"
    LABORATORY = "laboratory"

@dataclass
class SubjectClassification:
    subject: Subject
    confidence: float
    keywords_matched: List[str]
    reasoning: str

class SubjectClassifier:
    COMPOUND_KEYWORDS = {
        Subject.KDE: ['implement runtime', 'implement engine', 'implement seed', 'improve runtime', 'improve bootstrap'],
        Subject.LABORATORY: ['runtime ecu assessment', 'runtime ecu self-assessment', 'ecu self-assessment', 'ecu self-test', 'ecu observation', 'bootstrap investigation', 'test drive', 'operational test', 'violation audit', 'governance model'],
        Subject.PROJECT: ['implement feature', 'implement protocol', 'test feature', 'integration test']
    }
    SIMPLE_KEYWORDS = {
        Subject.KDE: ['runtime', 'bootstrap', 'engine', 'seed', 'governance', 'policy', 'capability', 'registry', 'kde', 'kdse', 'framework'],
        Subject.LABORATORY: ['self-assessment', 'self-test', 'self', 'laboratory', 'architecture', 'assessment', 'audit', 'routing', 'classification', 'procurement'],
        Subject.PROJECT: ['feature', 'protocol', 'library', 'bug', 'api', 'dnp3', 'atlas', 'ppc', 'librarian']
    }
    DEFAULT_SUBJECT = Subject.PROJECT

    def classify(self, description: str, title: str = "") -> SubjectClassification:
        combined = f"{description} {title}".lower()
        compound_score = {s: 0 for s in Subject}
        compound_matches = {s: [] for s in Subject}
        for subject, keywords in self.COMPOUND_KEYWORDS.items():
            for kw in keywords:
                if kw in combined:
                    compound_score[subject] += 3
                    compound_matches[subject].append(kw)
        simple_score = {s: 0 for s in Subject}
        simple_matches = {s: [] for s in Subject}
        for subject, keywords in self.SIMPLE_KEYWORDS.items():
            for kw in keywords:
                if kw in combined:
                    simple_score[subject] += 1
                    simple_matches[subject].append(kw)
        total_score = {s: compound_score[s] + simple_score[s] for s in Subject}
        all_matches = {s: compound_matches[s] + simple_matches[s] for s in Subject}
        if max(total_score.values()) > 0:
            subject = max(total_score, key=total_score.get)
            total = sum(total_score.values())
            confidence = total_score[subject] / total if total > 0 else 0.5
            matches = all_matches[subject]
        else:
            subject = self.DEFAULT_SUBJECT
            confidence = 0.5
            matches = []
        return SubjectClassification(subject=subject, confidence=confidence, keywords_matched=matches, reasoning=f"Keywords: {matches}" if matches else "No matches")
    
    def classify_from_id(self, artifact_id: str) -> Subject:
        id_upper = artifact_id.upper()
        if id_upper.startswith('KDE-'):
            return Subject.KDE
        elif id_upper.startswith('LAB-'):
            return Subject.LABORATORY
        return Subject.PROJECT
    
    def get_laboratory_path(self, subject: Subject) -> str:
        if subject in (Subject.KDE, Subject.LABORATORY):
            return ".kde/laboratory/"
        return "laboratory/"
    
    def get_id_prefix(self, subject: Subject, artifact_type: str) -> str:
        prefixes = {'investigation': 'INV', 'experiment': 'EXP', 'implementation': 'IMP', 'decision': 'TDR', 'review': 'REV', 'planning': 'PLAN', 'evidence': 'EVD', 'testing': 'TEST'}
        code = prefixes.get(artifact_type.lower(), 'INV')
        if subject == Subject.KDE:
            return f"KDE-{code}-"
        elif subject == Subject.LABORATORY:
            return f"LAB-{code}-"
        return f"PROJECT-{code}-"

class SubjectRouter:
    def __init__(self, kde_lab_path: str, project_lab_path: str):
        self.kde_lab_path = kde_lab_path
        self.project_lab_path = project_lab_path
        self.classifier = SubjectClassifier()
    
    def route(self, description: str, title: str = "") -> Dict:
        result = self.classifier.classify(description, title)
        subject = result.subject
        if subject in (Subject.KDE, Subject.LABORATORY):
            lab = self.kde_lab_path
            routing = f"KDE_LAB ({subject.value})"
        else:
            lab = self.project_lab_path
            routing = "PROJECT_LAB"
        return {'subject': subject, 'confidence': result.confidence, 'laboratory': lab, 'routing': routing, 'reasoning': result.reasoning}
