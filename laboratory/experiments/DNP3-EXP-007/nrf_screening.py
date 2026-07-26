#!/usr/bin/env python3
"""
NRF Article Screening System — PART 1: RELEVANCE CHECK & PART 2: CATEGORIZATION

This module implements the NRF Article Screening System to classify articles
from the NRF media monitoring spreadsheet.

Usage:
    python3 nrf_screening.py [input.xlsx] [--output results.csv]
"""

import re
import pandas as pd
from dataclasses import dataclass
from typing import Optional, Tuple, List
from pathlib import Path

# =============================================================================
# MANDATED ORGANIZATIONS (RULE 0)
# =============================================================================

MANDATED_ORGS = {
    # IHL - Institutes of Higher Learning
    "NUS", "NTU", "SMU", "SUTD", "SIT", "SUSS",
    
    # Hospitals / Clinicians
    "SingHealth", "SGH", "NUH", "TTSH", "CGH", "KTPH", "SKH", 
    "NTFH", "AH", "WH", "KKH", "NCCS", "NHCS", "SNEC", "IMH", "NDCS", "NSC",
    
    # Academic Medical Centres
    "SingHealth Duke-NUS", "NUHS", "NHG-LKCMedicine",
    
    # A*STAR
    "A*STAR", "ID Labs", "A*SRL", "BII", "BTI", "GIS", "IMCB", "SIgN", "SICS", "SIFBI",
    
    # National Research Infrastructure
    "TCOMS", "National Supercomputing Centre of Singapore", 
    "St John's Island National Marine Laboratory",
    
    # Research Institutes
    "COI BE-AM", "MBI", "IDMXS", "I-FIM", "CQT", "CSI", "ERI@N", "NEWRI", "SCELSE",
    
    # CREATE
    "CREATE", "BEARS", "CARES", "E2S2", "SHARE", "SEC", "SMART", 
    "TUM CREATE", "TSCP", "CNRS@CREATE"
}

# =============================================================================
# TOPIC MAPPINGS
# =============================================================================

# The 21 Framework Topics
TOPIC_MAPPINGS = {
    # MTC
    "manufacturing_material_science": ["manufacturing", "material science", "materials", 
                                       "graphene", "alloys", "nanomaterials", "superconductors",
                                       "semiconductor", "silicon", "photonics", "laser", "optics"],
    "robotics_industrial": ["industrial robot", "factory robot", "manufacturing robot", "automation"],
    "trade_connectivity": ["trade", "export", "import", "supply chain", "logistics", "shipping"],
    "electronics": ["electronics", "chip", "wafer", "mems", "semiconductor fabrication"],
    
    # HHP
    "gusto": ["gusto", "prenatal", "maternal", "child development", "childhood", "ADHD"],
    "precise": ["precise", "genomics", "genetic", "gene editing", "personalized medicine", 
                "pharmacogenomics", "rare disease"],
    "ageing": ["ageing", "aging", "longevity", "dementia", "alzheimer", "cognitive decline"],
    "prepare": ["prepare", "vaccine", "diagnostic", "therapeutic", "epidemic", "pandemic",
                "TB", "dengue", "influenza", "COVID", "HIV"],
    "general_health": ["health", "cancer", "cardiovascular", "diabetes", "neurological",
                       "eye health", "mental health", "cell therapy", "gene therapy", "medical device"],
    
    # USS
    "climate_change": ["climate change", "sea level", "greenhouse gas", "extreme weather",
                       "COP", "IPCC", "marine science", "coral reef", "ocean"],
    "sustainability": ["sustainability", "sustainable", "renewable", "solar", "wind",
                      "hydrogen", "EV", "electric vehicle", "decarbonisation", "circular economy"],
    "resource_resilience": ["food security", "agritech", "aquaculture", "alternative protein",
                            "food waste", "30-by-30", "water treatment", "membrane"],
    "built_environment": ["green construction", "urban liveability", "building", "drone inspection",
                          "sustainable concrete", "urban heat"],
    
    # SNDE
    "smart_nation": ["smart nation", "digital economy", "digital transformation"],
    "cybersecurity": ["cybersecurity", "security", "cryptography", "privacy", "data protection"],
    "communications": ["5G", "6G", "wireless", "quantum communication", "cloud"],
    "quantum": ["quantum", "quantum computing", "quantum network", "supercomputing", "HPC", "GPU"],
    "robotics": ["robot", "humanoid", "service robot", "autonomous"]
}

# Cross-cutting topics
CROSS_CUTTING = {
    "innovation_enterprise": ["AI chip", "tech war", "nvidia", "amd", "openai", "deepseek",
                             "innovation", "startup", "enterprise"],
    "manpower": ["talent", "training", "scholarship", "PhD", "STEM", "researcher career",
                 "workforce", "skills"],
    "academic_research": ["research grant", "corporate lab", "consortium", "fellowship",
                          "research collaboration"]
}

# Category definitions
CATEGORIES = ["MTC", "HHP", "USS", "SNDE"]

# Category to domain mapping
DOMAIN_TOPICS = {
    "MTC": ["manufacturing_material_science", "robotics_industrial", "trade_connectivity", "electronics"],
    "HHP": ["gusto", "precise", "ageing", "prepare", "general_health"],
    "USS": ["climate_change", "sustainability", "resource_resilience", "built_environment"],
    "SNDE": ["smart_nation", "cybersecurity", "communications", "quantum", "robotics"]
}

# =============================================================================
# DATA CLASSES
# =============================================================================

@dataclass
class ScreeningResult:
    """Result of NRF article screening."""
    classification: str  # RELEVANT | IRRELEVANT
    mandate: str  # Yes — org | No
    domain: str  # MTC | HHP | USS | SNDE | NONE
    lf_topic: str  # One of 21 topics | NONE
    gate_failed: str  # 0 | 1 | 2 | 3 | None
    reasoning: str
    confidence: str  # HIGH | MEDIUM | LOW
    sg_entity: str  # Yes — entity | No
    second_pass: str  # Yes | No
    category: str = ""  # Part 2: MTC | HHP | USS | SNDE
    agrees_with_part1: str = ""  # Yes | No

# =============================================================================
# SCREENING ENGINE
# =============================================================================

class NRFScreeningEngine:
    """NRF Article Screening Engine - Parts 1 and 2."""
    
    def __init__(self):
        self.mandated_orgs = MANDATED_ORGS
        self.topic_mappings = TOPIC_MAPPINGS
        self.cross_cutting = CROSS_CUTTING
        self.categories = CATEGORIES
    
    def screen(self, title: str, content: str = "") -> ScreeningResult:
        """
        Screen an article through Part 1 (Relevance) and Part 2 (Categorization).
        
        Args:
            title: Article title
            content: Article content (may be empty)
        
        Returns:
            ScreeningResult with full classification
        """
        # Part 1: Relevance Check
        result = self.part1_relevance(title, content)
        
        # Part 2: Categorization (only if RELEVANT)
        if result.classification == "RELEVANT":
            cat_result = self.part2_categorize(title, content, result.domain, result.lf_topic)
            result.category = cat_result["category"]
            result.agrees_with_part1 = cat_result["agrees"]
        
        return result
    
    def part1_relevance(self, title: str, content: str) -> ScreeningResult:
        """Part 1: Relevance Check - Rules 0 and Gates 0-3."""
        
        text = f"{title} {content}".lower()
        
        # RULE 0: Standing Order Override
        org_result = self.check_mandated_orgs(title, content)
        if org_result:
            return ScreeningResult(
                classification="RELEVANT",
                mandate=f"Yes — {org_result}",
                domain="NONE",
                lf_topic="NONE",
                gate_failed="None",
                reasoning=f"Mandated organization {org_result} found in article",
                confidence="HIGH",
                sg_entity=f"Yes — {org_result}",
                second_pass="No"
            )
        
        # GATE 0: Article Existence
        gate0_result = self.gate0_article_exists(title, content)
        if gate0_result:
            return gate0_result
        
        # GATE 1: Framework Fit
        gate1_result = self.gate1_framework_fit(title, content)
        if gate1_result:
            return gate1_result
        
        # GATE 2: Portability
        gate2_result = self.gate2_portability(title, content)
        if gate2_result:
            return gate2_result
        
        # GATE 3: Research Substance
        gate3_result = self.gate3_substance(title, content)
        if gate3_result:
            return gate3_result
        
        # Default: Pass with HIGH confidence
        domain, topic = self.identify_domain_topic(title, content)
        return ScreeningResult(
            classification="RELEVANT",
            mandate="No",
            domain=domain,
            lf_topic=topic,
            gate_failed="None",
            reasoning="Article passes all gates",
            confidence="HIGH",
            sg_entity=self.detect_sg_entity(title, content),
            second_pass="No"
        )
    
    def check_mandated_orgs(self, title: str, content: str) -> Optional[str]:
        """Rule 0: Check for mandated organizations."""
        text = f"{title} {content}"
        
        # Case-sensitive match for acronyms to avoid false positives
        for org in self.mandated_orgs:
            # Check for whole word match
            pattern = r'\b' + re.escape(org) + r'\b'
            if re.search(pattern, text):
                return org
        
        return None
    
    def gate0_article_exists(self, title: str, content: str) -> Optional[ScreeningResult]:
        """Gate 0: Check if there's an actual article here."""
        
        # Paywall stub check
        if not content or len(content.strip()) < 50:
            return ScreeningResult(
                classification="IRRELEVANT",
                mandate="No",
                domain="NONE",
                lf_topic="NONE",
                gate_failed="0",
                reasoning="Paywall stub or insufficient content",
                confidence="HIGH",
                sg_entity=self.detect_sg_entity(title, content),
                second_pass="No"
            )
        
        # Roundup check
        roundup_patterns = [
            r"latest news releases",
            r"while you were sleeping",
            r"stories you might have missed",
            r"morning briefing",
            r"evening briefing",
            r"daily briefing",
            r"in case you missed",
            r"news digest",
            r"weekly round-up",
            r"call for papers",
            r"media advisory",
            r"obituar",
            r"in memoriam"
        ]
        
        text_lower = f"{title} {content}".lower()
        for pattern in roundup_patterns:
            if re.search(pattern, text_lower):
                return ScreeningResult(
                    classification="IRRELEVANT",
                    mandate="No",
                    domain="NONE",
                    lf_topic="NONE",
                    gate_failed="0",
                    reasoning=f"Roundup or diary entry detected: {pattern}",
                    confidence="HIGH",
                    sg_entity=self.detect_sg_entity(title, content),
                    second_pass="No"
                )
        
        return None
    
    def gate1_framework_fit(self, title: str, content: str) -> Optional[ScreeningResult]:
        """Gate 1: Check if article fits the Listening Framework."""
        
        text = f"{title} {content}".lower()
        sg_entity = self.detect_sg_entity(title, content)
        
        # Check 21 topics
        matched_topics = []
        for category, topics in DOMAIN_TOPICS.items():
            for topic in topics:
                if topic in self.topic_mappings:
                    for keyword in self.topic_mappings[topic]:
                        if keyword.lower() in text:
                            matched_topics.append((category, topic, keyword))
                            break
        
        # Check cross-cutting topics
        for topic_name, keywords in self.cross_cutting.items():
            for keyword in keywords:
                if keyword.lower() in text:
                    matched_topics.append(("CROSS", topic_name, keyword))
                    break
        
        if not matched_topics:
            # Check for exclusion topics
            exclusion_topics = [
                "planetary science", "astronomy", "palaeontology", "paleontology",
                "geology", "origin of life", "education research", "pedagogy",
                "organizational psychology", "HR", "workplace behaviour",
                "sports science", "veterinary", "animal welfare",
                "plant biology", "arts", "humanities", "social science"
            ]
            
            for exclusion in exclusion_topics:
                if exclusion in text:
                    return ScreeningResult(
                        classification="IRRELEVANT",
                        mandate="No",
                        domain="NONE",
                        lf_topic="NONE",
                        gate_failed="1",
                        reasoning=f"Article outside framework: {exclusion} topic",
                        confidence="HIGH",
                        sg_entity=sg_entity,
                        second_pass="No"
                    )
            
            return ScreeningResult(
                classification="IRRELEVANT",
                mandate="No",
                domain="NONE",
                lf_topic="NONE",
                gate_failed="1",
                reasoning="Article does not fit any of 21 framework topics",
                confidence="HIGH",
                sg_entity=sg_entity,
                second_pass="No"
            )
        
        # Return the best match
        # Prefer non-CROSS matches for domain
        non_cross = [m for m in matched_topics if m[0] != "CROSS"]
        if non_cross:
            domain, topic, kw = non_cross[0]
        else:
            domain, topic, kw = matched_topics[0]
            domain = "CROSS"
        
        return None  # Continue to next gate
    
    def gate2_portability(self, title: str, content: str) -> ScreeningResult:
        """Gate 2: Check if findings are portable."""
        
        text = f"{title} {content}".lower()
        sg_entity = self.detect_sg_entity(title, content)
        
        # Global frameworks always pass
        global_frameworks = ["cop", "ipcc", "who", "un ", "united nations", "global health"]
        for framework in global_frameworks:
            if framework in text:
                return None  # Pass
        
        # Check if location is load-bearing
        se_asia = ["singapore", "southeast asia", "sea", "asean", "malaysia", "indonesia", 
                   "thailand", "vietnam", "philippines", "mekong", "borneo", "sumatra"]
        
        has_sg = any(loc in text for loc in se_asia)
        
        # Location-heavy patterns
        location_patterns = [
            r"\bafrica\b", r"\beurope\b", r"\bamerica\b", r"\bchina\b", r"\bindia\b",
            r"\bjapan\b", r"\bkorea\b", r"\baustralia\b", r"\buk\b", r"\bbritain\b"
        ]
        
        has_foreign = any(re.search(p, text) for p in location_patterns)
        
        # If location mentioned but no SG, might be IRRELEVANT
        if has_foreign and not has_sg:
            # Check if it's about SE Asia context
            if any(loc in text for loc in se_asia[1:]):  # Exclude Singapore
                return None  # Pass - SE Asia context
        
        return None  # Pass
    
    def gate3_substance(self, title: str, content: str) -> ScreeningResult:
        """Gate 3: Check for research or strategic substance."""
        
        text = f"{title} {content}".lower()
        sg_entity = self.detect_sg_entity(title, content)
        domain, topic = self.identify_domain_topic(title, content)
        
        # Research substance indicators
        research_indicators = [
            "study", "research", "trial", "experiment", "discovery", 
            "breakthrough", "method", "platform", "clinical", "deployment",
            "findings", "results", "analysis", "published", "journal",
            "preprint", "patent", "grant", "funding"
        ]
        
        innovation_indicators = [
            "R&D", "innovation", "development", "technology transfer",
            "collaboration", "partnership", "facility", "center", "centre",
            "lab", "laboratory", "research institute"
        ]
        
        has_research = any(ind in text for ind in research_indicators)
        has_innovation = any(ind in text for ind in innovation_indicators)
        
        # Financial-only patterns (fail)
        financial_only = [
            r"\$[\d,]+ (?:million|billion) (?:series|fund|investment|valuation)",
            r"raises? \$\d+",
            r"IPO", r"stock listing", r"merger", r"acquisition",
            r"earnings", r"revenue", r"quarterly results"
        ]
        
        for pattern in financial_only:
            if re.search(pattern, text, re.IGNORECASE):
                # Check if it's about R&D context
                if not (has_research or has_innovation):
                    return ScreeningResult(
                        classification="IRRELEVANT",
                        mandate="No",
                        domain=domain if domain else "NONE",
                        lf_topic=topic if topic else "NONE",
                        gate_failed="3",
                        reasoning="Pure financial event with no R&D content",
                        confidence="HIGH",
                        sg_entity=sg_entity,
                        second_pass="No"
                    )
        
        return None  # Pass - has substance
    
    def identify_domain_topic(self, title: str, content: str) -> Tuple[str, str]:
        """Identify the primary domain and topic from text."""
        
        text = f"{title} {content}".lower()
        
        # Check topics in priority order
        for category, topics in DOMAIN_TOPICS.items():
            for topic in topics:
                if topic in self.topic_mappings:
                    for keyword in self.topic_mappings[topic]:
                        if keyword.lower() in text:
                            return (category, topic)
        
        # Default
        return ("NONE", "NONE")
    
    def part2_categorize(self, title: str, content: str, initial_domain: str, 
                         lf_topic: str) -> dict:
        """Part 2: Categorization into MTC, HHP, USS, or SNDE."""
        
        text = f"{title} {content}".lower()
        domain = initial_domain if initial_domain and initial_domain != "NONE" else self.identify_domain(title, content)
        
        # Rule 1: All nuclear → USS
        if "nuclear" in text:
            return {
                "category": "USS",
                "agrees": "No — overturned from " + domain if domain != "USS" and domain != "NONE" else "Yes",
                "reasoning": "Nuclear energy/policy automatically categorized as USS"
            }
        
        # Rule 2: Nvidia/AI/Quantum → SNDE (default)
        ai_quantum = ["nvidia", "amd", "qualcomm", "intel ai", "ai chip", 
                      "quantum computing", "quantum network", "openai", "deepseek"]
        
        is_ai_quantum = any(keyword in text for keyword in ai_quantum)
        
        if is_ai_quantum:
            # Rule 3: Check if domain override applies
            override = self.check_domain_override(title, content)
            if override:
                return {
                    "category": override,
                    "agrees": "No — overturned from SNDE",
                    "reasoning": f"AI/quantum applied to {override} domain (Rule 3)"
                }
            
            return {
                "category": "SNDE",
                "agrees": "Yes" if domain == "SNDE" else "No — overturned from " + domain,
                "reasoning": "AI/quantum technology as primary subject defaults to SNDE"
            }
        
        # Rule 3: Domain override - what the work does
        override = self.check_domain_override(title, content)
        if override:
            return {
                "category": override,
                "agrees": "Yes" if domain == override else "No — overturned from " + domain,
                "reasoning": f"Domain override applied: {override} (Rule 3)"
            }
        
        # Rule 4: Energy/sustainability → USS
        energy_topics = ["battery", "energy storage", "solar", "photovoltaic", 
                        "wind energy", "hydrogen", "geothermal", "renewable",
                        "EV", "electric vehicle", "sustainable aviation"]
        
        if any(topic in text for topic in energy_topics):
            return {
                "category": "USS",
                "agrees": "Yes" if domain == "USS" else "No — overturned from " + domain,
                "reasoning": "Energy/sustainability topic categorized as USS (Rule 4)"
            }
        
        # Default: use identified domain
        if domain in CATEGORIES:
            return {
                "category": domain,
                "agrees": "Yes",
                "reasoning": f"Article categorized as {domain} based on primary subject"
            }
        
        # Fallback: SNDE for tech-heavy content
        tech_keywords = ["AI", "digital", "computer", "software", "data", "cyber"]
        if any(kw in text for kw in tech_keywords):
            return {
                "category": "SNDE",
                "agrees": "No — overturned from " + domain if domain != "NONE" else "Yes",
                "reasoning": "Technology-focused content categorized as SNDE"
            }
        
        # Final fallback
        return {
            "category": domain if domain in CATEGORIES else "SNDE",
            "agrees": "Yes",
            "reasoning": "Default categorization"
        }
    
    def check_domain_override(self, title: str, content: str) -> Optional[str]:
        """Check for Rule 3 domain overrides."""
        
        text = f"{title} {content}".lower()
        
        # AI for specific domains
        if "AI for" in text or "artificial intelligence for" in text:
            if any(health in text for health in ["cancer", "diagnosis", "health", "medical", "clinical"]):
                return "HHP"
            if any(mat in text for mat in ["material", "manufacturing", "chip", "semiconductor"]):
                return "MTC"
            if "food" in text or "waste" in text:
                return "USS"
        
        # Quantum for manufacturing
        if "quantum" in text and any(mfg in text for mfg in ["manufacturing", "inspection", "quality"]):
            return "MTC"
        
        # Quantum for research/science
        if "quantum" in text and any(sci in text for sci in ["error correction", "algorithm", "computing"]):
            return "SNDE"
        
        return None
    
    def identify_domain(self, title: str, content: str) -> str:
        """Identify the primary domain from text."""
        
        domain, _ = self.identify_domain_topic(title, content)
        return domain if domain else "SNDE"
    
    def detect_sg_entity(self, title: str, content: str) -> str:
        """Detect if a Singapore entity is mentioned."""
        
        text = f"{title} {content}"
        
        sg_entities = [
            "Singapore", "Singaporean", "NUS", "NTU", "SMU", "SUTD",
            "SingHealth", "A*STAR", "NRF", "TCOMS", "IMCB", "CQT"
        ]
        
        for entity in sg_entities:
            if entity in text:
                return f"Yes — {entity}"
        
        return "No"
    
    def format_output(self, result: ScreeningResult) -> str:
        """Format the screening result for output."""
        
        output = []
        output.append("=" * 60)
        output.append("NRF ARTICLE SCREENING RESULTS")
        output.append("=" * 60)
        
        output.append(f"\nCLASSIFICATION: {result.classification}")
        output.append(f"MANDATE: {result.mandate}")
        output.append(f"DOMAIN: {result.domain}")
        output.append(f"LF TOPIC: {result.lf_topic}")
        output.append(f"GATE FAILED: {result.gate_failed}")
        output.append(f"REASONING: {result.reasoning}")
        output.append(f"CONFIDENCE LEVEL: {result.confidence}")
        output.append(f"SINGAPORE ENTITY MENTIONED: {result.sg_entity}")
        output.append(f"REQUIRES SECOND PASS: {result.second_pass}")
        
        if result.category:
            output.append("\n" + "-" * 40)
            output.append("PART 2: CATEGORIZATION")
            output.append("-" * 40)
            output.append(f"CATEGORY: {result.category}")
            output.append(f"AGREES WITH PART 1: {result.agrees_with_part1}")
        
        output.append("=" * 60)
        
        return "\n".join(output)


def load_articles(filepath: str) -> List[dict]:
    """Load articles from Excel file."""
    
    xlsx = pd.ExcelFile(filepath)
    
    # Try different sheets
    for sheet_name in ['Consolidated SL - Jan', 'Data', 'Template']:
        if sheet_name in xlsx.sheet_names:
            df = pd.read_excel(xlsx, sheet_name=sheet_name, header=None)
            # Extract headlines and content
            articles = []
            for idx in range(1, min(len(df), 50)):  # Process first 50 articles
                row = df.iloc[idx]
                headline = str(row.iloc[6]) if len(row) > 6 else ""
                content = str(row.iloc[8]) if len(row) > 8 else ""
                outlet = str(row.iloc[5]) if len(row) > 5 else ""
                date = str(row.iloc[3]) if len(row) > 3 else ""
                
                if headline and headline != 'nan':
                    articles.append({
                        "headline": headline,
                        "content": content,
                        "outlet": outlet,
                        "date": date
                    })
            
            if articles:
                return articles
    
    return []


def main():
    """Main function."""
    import argparse
    
    parser = argparse.ArgumentParser(description="NRF Article Screening System")
    parser.add_argument("input", nargs="?", default="/tmp/spreadsheet.xlsx",
                       help="Input Excel file")
    parser.add_argument("--output", "-o", default="screening_results.csv",
                       help="Output CSV file")
    parser.add_argument("--limit", "-l", type=int, default=20,
                       help="Limit number of articles to process")
    parser.add_argument("--verbose", "-v", action="store_true",
                       help="Verbose output")
    
    args = parser.parse_args()
    
    # Initialize engine
    engine = NRFScreeningEngine()
    
    # Load articles
    print(f"Loading articles from {args.input}...")
    articles = load_articles(args.input)
    
    if not articles:
        print("No articles found!")
        return
    
    print(f"Found {len(articles)} articles, processing first {args.limit}...")
    
    # Process articles
    results = []
    for i, article in enumerate(articles[:args.limit]):
        result = engine.screen(article["headline"], article["content"])
        result.outlet = article["outlet"]
        result.date = article["date"]
        results.append(result)
        
        if args.verbose:
            print(f"\n[{i+1}/{min(args.limit, len(articles))}] {article['headline'][:60]}...")
            print(engine.format_output(result))
    
    # Summary
    print("\n" + "=" * 60)
    print("SCREENING SUMMARY")
    print("=" * 60)
    
    relevant = sum(1 for r in results if r.classification == "RELEVANT")
    irrelevant = sum(1 for r in results if r.classification == "IRRELEVANT")
    second_pass = sum(1 for r in results if r.second_pass == "Yes")
    
    print(f"\nTotal processed: {len(results)}")
    print(f"RELEVANT: {relevant} ({relevant/len(results)*100:.1f}%)")
    print(f"IRRELEVANT: {irrelevant} ({irrelevant/len(results)*100:.1f}%)")
    print(f"Requires second pass: {second_pass}")
    
    # Category breakdown
    print("\nCategory breakdown (RELEVANT only):")
    cat_counts = {}
    for r in results:
        if r.classification == "RELEVANT" and r.category:
            cat_counts[r.category] = cat_counts.get(r.category, 0) + 1
    
    for cat, count in sorted(cat_counts.items()):
        print(f"  {cat}: {count}")
    
    # Gate failures
    print("\nGate failure breakdown:")
    gate_counts = {}
    for r in results:
        if r.gate_failed and r.gate_failed != "None":
            gate_counts[r.gate_failed] = gate_counts.get(r.gate_failed, 0) + 1
    
    for gate, count in sorted(gate_counts.items()):
        print(f"  Gate {gate}: {count}")
    
    # Save to CSV
    df_results = pd.DataFrame([{
        "Headline": articles[i]["headline"],
        "Outlet": r.outlet,
        "Date": r.date,
        "Classification": r.classification,
        "Mandate": r.mandate,
        "Domain": r.domain,
        "LF_Topic": r.lf_topic,
        "Gate_Failed": r.gate_failed,
        "Reasoning": r.reasoning,
        "Confidence": r.confidence,
        "SG_Entity": r.sg_entity,
        "Second_Pass": r.second_pass,
        "Category": r.category,
        "Agrees_Part1": r.agrees_with_part1
    } for i, r in enumerate(results)])
    
    df_results.to_csv(args.output, index=False)
    print(f"\nResults saved to {args.output}")


if __name__ == "__main__":
    main()
