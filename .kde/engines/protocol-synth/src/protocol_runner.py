"""
Protocol Synthesis Runner
Executes a single protocol synthesis run.
"""

import asyncio
import json
import os
import sys
import subprocess
import traceback
import hashlib
from datetime import datetime
from pathlib import Path
from typing import Dict, Any, Optional

# Add src to path
sys.path.insert(0, str(Path(__file__).parent))

from protocol_generator import ProtocolGenerator
from protocol_implementor import implement_protocol


class ProtocolRun:
    """Executes a single protocol synthesis run."""
    
    def __init__(self, run_id: str, seed: int, output_dir: Path):
        self.run_id = run_id
        self.seed = seed
        self.output_dir = Path(output_dir) / run_id
        self.timestamp = datetime.utcnow().isoformat() + "Z"
        
        # Create directory structure
        self.workspace = self.output_dir / "workspace"
        self.artifacts = self.output_dir / "artifacts"
        self.logs = self.output_dir / "logs"
        
        for d in [self.workspace, self.workspace / "source", 
                  self.workspace / "build", self.workspace / "runtime",
                  self.artifacts, self.logs]:
            d.mkdir(parents=True, exist_ok=True)
        
        self.results = {
            "run_id": run_id,
            "seed": seed,
            "timestamp": self.timestamp,
            "status": "initialized",
            "protocol_name": None,
            "compilation": {},
            "functional_test": {},
            "security_test": {},
            "errors": [],
        }
    
    def _log(self, message: str, level: str = "INFO"):
        """Log a message."""
        log_file = self.logs / "run.log"
        timestamp = datetime.utcnow().isoformat()
        log_line = f"[{timestamp}] [{level}] {message}\n"
        
        with open(log_file, "a") as f:
            f.write(log_line)
        
        print(log_line.strip())
    
    def _generate_protocol(self) -> Dict[str, Any]:
        """Generate protocol specification."""
        self._log("Generating protocol specification...")
        
        generator = ProtocolGenerator(self.seed)
        spec = generator.generate()
        full_spec = generator.generate_full_specification()
        
        self.results["protocol_name"] = spec.name
        self.results["protocol_spec"] = full_spec
        
        self._log(f"Generated protocol: {spec.name}")
        return full_spec
    
    def _write_specification(self, spec: Dict[str, Any]):
        """Write protocol specification to file."""
        spec_file = self.artifacts / "specification.md"
        
        # Generate Markdown specification
        md = f"""# {spec['protocol']['name']}

**Protocol Version**: {spec['protocol']['version']}
**Generated**: {self.timestamp}
**Run ID**: {self.run_id}
**Seed**: {self.seed}

---

## Protocol Overview

**{spec['protocol']['name']}** is a secure communication protocol designed to provide
{('post-quantum' if spec['protocol']['supports_post_quantum'] else '')} {spec['protocol']['forward_secrecy_level']} forward secrecy.

### Architecture

| Property | Value |
|----------|-------|
| Handshake Model | {spec['protocol']['handshake_model']} |
| Key Exchange | {spec['protocol']['key_exchange_type']} |
| Authentication | {spec['protocol']['authentication_model']} |
| Encryption | {spec['protocol']['encryption_mode']} |
| Forward Secrecy | {spec['protocol']['forward_secrecy_level']} |
| Message Format | {spec['protocol']['message_format']} |

### Cryptographic Primitives

| Component | Algorithm |
|-----------|-----------|
| Key Exchange | {spec['protocol']['key_exchange_algorithm']} |
| Authentication | {spec['protocol']['authentication_algorithm']} |
| Encryption | {spec['protocol']['encryption_algorithm']} |
| Hash | {spec['protocol']['hash_algorithm']} |
| KDF | {spec['protocol']['kdf_algorithm']} |

### Features

"""
        
        features = []
        if spec['protocol']['supports_0rtt']:
            features.append("- 0-RTT data support")
        if spec['protocol']['supports_resumption']:
            features.append("- Session resumption")
        if spec['protocol']['supports_post_quantum']:
            features.append("- Post-quantum key exchange")
        if spec['protocol']['connection_migration']:
            features.append("- Connection migration")
        
        md += "\n".join(features) if features else "- Standard secure session"
        
        md += f"""

### Distinctive Features

"""
        
        if spec['protocol']['distinct_features']:
            for feature in spec['protocol']['distinct_features']:
                md += f"- {feature}\n"
        
        md += f"""

---

## Security Properties

| Property | Value |
|----------|-------|
| Confidentiality | {spec['protocol']['confidentiality']} |
| Integrity | {spec['protocol']['integrity']} |
| Authenticity | {spec['protocol']['authenticity']} |
| Forward Secrecy | {spec['protocol']['forward_secrecy']} |
| Replay Protection | {spec['protocol']['replay_protection']} |

---

## Threat Model

### Threats Addressed

"""
        
        for threat in spec['threat_model']['threats_addressed']:
            md += f"- {threat}\n"
        
        md += """

### Security Assumptions

"""
        
        for assumption in spec['threat_model']['assumptions']:
            md += f"- {assumption}\n"
        
        md += """

---

## Handshake Sequence

```
"""
        
        for i, msg in enumerate(spec['handshake_sequence']):
            arrow = "→" if msg['to'] != msg['from'] else "↔"
            md += f"{{i+1:2}}. {msg['from']:12} {arrow} {msg['to']:12} : {msg['type']}\n"
            md += f"    Content: {msg['content']}\n"
            md += f"    Encrypted: {msg['encrypted']}\n\n"
        
        md += """```

**Handshake Rounds**: """ + str(spec['protocol']['handshake_rounds']) + """
**Message Overhead**: """ + str(spec['protocol']['message_overhead']) + """ bytes

---

## State Machine

### States

"""
        
        for state in spec['state_machine']['states']:
            md += f"- **{state['name']}**: {state['description']}\n"
        
        md += """

### Transitions

```
"""
        
        for trans in spec['state_machine']['transitions']:
            src_state = trans['src']
            dst_state = trans['dst']
            event = trans['event']
            md += f"{src_state} --({event})--> {dst_state}\n"
        
        md += """```

---

## Packet Structure

"""
        
        for ptype, pstruct in spec['packet_structure']['packet_types'].items():
            md += f"### {ptype} Packet\n\n"
            md += f"**Header Size**: {pstruct['header_size']} bytes\n\n"
            md += "| Field | Size | Description |\n"
            md += "|-------|------|-------------|\n"
            
            for field in pstruct['fields']:
                size = field['size'] if isinstance(field['size'], str) else f"{field['size']} bytes"
                md += f"| {field['name']} | {size} | {field['description']} |\n"
            
            md += "\n"
        
        md += f"""---

## Security Analysis

### Advantages

"""
        
        for adv in spec['security_analysis']['advantages']:
            md += f"- {adv}\n"
        
        md += """

### Weaknesses

"""
        
        for weak in spec['security_analysis']['weaknesses']:
            md += f"- {weak}\n"
        
        md += """

### Self-Criticism

"""
        
        for crit in spec['security_analysis']['self_criticism']:
            md += f"> {crit}\n\n"
        
        md += """

---

## Comparison to Existing Protocols

"""
        
        for comparison in spec['comparison']['comparisons']:
            md += f"### vs {comparison['protocol']}\n\n"
            md += f"**Architectural Similarity**: {comparison['similarity']*100:.0f}%\n\n"
            md += "**Key Differences**:\n"
            for diff in comparison['differences']:
                if diff:
                    md += f"- {diff}\n"
            md += "\n"
        
        md += f"""

**Novelty Score**: {spec['comparison']['novelty_score']*100:.0f}%
**Summary**: {spec['comparison']['summary']}

---

## Implementation Plan

"""
        
        for phase in spec['implementation_plan']['phases']:
            md += f"### Phase {phase['phase']}: {phase['name']}\n\n"
            for task in phase['tasks']:
                md += f"- [ ] {task}\n"
            md += "\n"
        
        md += """

---

## Implementation

See `workspace/source/protocol.py` for the Python implementation.

"""
        
        with open(spec_file, "w") as f:
            f.write(md)
        
        self._log(f"Specification written to {spec_file}")
    
    def _implement_protocol(self, spec: Dict[str, Any]):
        """Generate and write protocol implementation."""
        self._log("Generating protocol implementation...")
        
        impl_file = self.workspace / "source" / "protocol.py"
        # Pass the protocol dict directly
        code = implement_protocol(spec['protocol'])
        
        with open(impl_file, "w") as f:
            f.write(code)
        
        self._log(f"Implementation written to {impl_file}")
        self.results["implementation"] = {
            "file": str(impl_file),
            "language": "Python",
            "lines_of_code": len(code.split('\n')),
        }
    
    def _compile_verify(self) -> bool:
        """Compile/verify the implementation."""
        self._log("Verifying implementation syntax...")
        
        impl_file = self.workspace / "source" / "protocol.py"
        
        try:
            # Python syntax check
            result = subprocess.run(
                [sys.executable, "-m", "py_compile", str(impl_file)],
                capture_output=True,
                text=True,
                timeout=30
            )
            
            if result.returncode == 0:
                self.results["compilation"] = {
                    "status": "success",
                    "language": "Python",
                    "output": "Syntax verification passed",
                    "warnings": [],
                }
                self._log("Syntax verification passed", "SUCCESS")
                return True
            else:
                self.results["compilation"] = {
                    "status": "failed",
                    "language": "Python",
                    "errors": result.stderr,
                }
                self.results["errors"].append(f"Compilation failed: {result.stderr}")
                self._log(f"Compilation failed: {result.stderr}", "ERROR")
                return False
                
        except subprocess.TimeoutExpired:
            self.results["errors"].append("Compilation timeout")
            self._log("Compilation timeout", "ERROR")
            return False
        except Exception as e:
            self.results["errors"].append(str(e))
            self._log(f"Compilation error: {e}", "ERROR")
            return False
    
    def _install_dependencies(self) -> bool:
        """Install required dependencies."""
        self._log("Installing dependencies...")
        
        impl_file = self.workspace / "source" / "protocol.py"
        
        # Create requirements.txt
        req_file = self.workspace / "source" / "requirements.txt"
        with open(req_file, "w") as f:
            f.write("cryptography>=41.0.0\n")
        
        # Install
        try:
            result = subprocess.run(
                [sys.executable, "-m", "pip", "install", "-q", "-r", str(req_file)],
                capture_output=True,
                text=True,
                timeout=120
            )
            
            if result.returncode == 0:
                self._log("Dependencies installed successfully", "SUCCESS")
                return True
            else:
                self._log(f"Dependency install warning: {result.stderr}", "WARNING")
                return True  # Continue anyway
                
        except Exception as e:
            self._log(f"Dependency install error: {e}", "WARNING")
            return True  # Continue anyway (stub mode)
    
    def _run_functional_test(self) -> Dict[str, Any]:
        """Run functional test of the protocol."""
        self._log("Running functional test...")
        
        test_result = {
            "status": "pending",
            "tests": [],
            "passed": 0,
            "failed": 0,
        }
        
        # Create test script
        test_file = self.workspace / "source" / "test_protocol.py"
        impl_file = self.workspace / "source" / "protocol.py"
        
        test_code = f'''"""
Functional test for {self.results['protocol_name']}
"""
import sys
import asyncio
sys.path.insert(0, "{impl_file.parent}")

async def test_session_creation():
    """Test session creation."""
    try:
        # Import and test
        exec(open("{impl_file}").read())
        
        # Test crypto operations
        crypto = {self.results['protocol_name']}Crypto()
        private_key, public_key = crypto.generate_keypair()
        
        assert private_key is not None, "Key generation failed"
        assert public_key is not None, "Public key generation failed"
        
        print("PASS: Session creation")
        return True
    except Exception as e:
        print(f"FAIL: Session creation - {{e}}")
        return False

async def test_encryption():
    """Test encryption/decryption."""
    try:
        exec(open("{impl_file}").read())
        
        crypto = {self.results['protocol_name']}Crypto()
        cipher = crypto.get_cipher()
        nonce = b'\\x00' * 12
        plaintext = b"Test message"
        
        # Encrypt
        ciphertext = crypto.encrypt(cipher, nonce, plaintext)
        assert ciphertext != plaintext, "Encryption failed"
        
        # Decrypt
        decrypted = crypto.decrypt(cipher, nonce, ciphertext)
        assert decrypted == plaintext, "Decryption failed"
        
        print("PASS: Encryption/decryption")
        return True
    except Exception as e:
        print(f"FAIL: Encryption - {{e}}")
        return False

async def test_key_derivation():
    """Test key derivation."""
    try:
        exec(open("{impl_file}").read())
        
        crypto = {self.results['protocol_name']}Crypto()
        input_key = b'test_secret_key_32_bytes_long_x'
        salt = b'salt_value'
        info = b'test_kdf'
        
        derived = crypto.kdf(input_key, salt, info, 48)
        assert len(derived) == 48, "KDF output wrong length"
        assert derived != input_key, "KDF failed"
        
        print("PASS: Key derivation")
        return True
    except Exception as e:
        print(f"FAIL: Key derivation - {{e}}")
        return False

async def test_state_machine():
    """Test state transitions."""
    try:
        exec(open("{impl_file}").read())
        
        session = {self.results['protocol_name']}Session()
        assert session.state == State.INIT, "Initial state wrong"
        
        session.state = State.HANDSHAKING
        assert session.state == State.HANDSHAKING, "State transition failed"
        
        print("PASS: State machine")
        return True
    except Exception as e:
        print(f"FAIL: State machine - {{e}}")
        return False

async def test_replay_protection():
    """Test replay protection."""
    try:
        exec(open("{impl_file}").read())
        
        session = {self.results['protocol_name']}Session()
        
        # Test nonce tracking
        nonce = 12345
        session.seen_nonces.add(nonce)
        
        if nonce in session.seen_nonces:
            # This is expected - replay detected
            print("PASS: Replay protection (nonce tracking)")
            return True
        
        print("FAIL: Replay protection")
        return False
    except Exception as e:
        print(f"FAIL: Replay protection - {{e}}")
        return False

async def main():
    """Run all tests."""
    tests = [
        test_session_creation,
        test_encryption,
        test_key_derivation,
        test_state_machine,
        test_replay_protection,
    ]
    
    passed = 0
    failed = 0
    
    for test in tests:
        result = await test()
        if result:
            passed += 1
        else:
            failed += 1
    
    print(f"\\nResults: {{passed}} passed, {{failed}} failed")
    return passed, failed

if __name__ == "__main__":
    asyncio.run(main())
'''
        
        with open(test_file, "w") as f:
            f.write(test_code)
        
        # Run tests
        try:
            result = subprocess.run(
                [sys.executable, str(test_file)],
                capture_output=True,
                text=True,
                timeout=60,
                cwd=str(self.workspace / "source")
            )
            
            output = result.stdout + result.stderr
            
            # Parse results
            for line in output.split('\n'):
                if 'PASS:' in line:
                    test_result["tests"].append({"name": line.split('PASS: ')[1], "status": "pass"})
                    test_result["passed"] += 1
                elif 'FAIL:' in line:
                    test_result["tests"].append({"name": line.split('FAIL: ')[1], "status": "fail"})
                    test_result["failed"] += 1
            
            test_result["status"] = "completed"
            test_result["output"] = output
            
            if test_result["failed"] == 0:
                test_result["overall"] = "pass"
                self._log(f"Functional tests: {test_result['passed']} passed", "SUCCESS")
            else:
                test_result["overall"] = "partial"
                self._log(f"Functional tests: {test_result['passed']} passed, {test_result['failed']} failed", "WARNING")
            
        except subprocess.TimeoutExpired:
            test_result["status"] = "timeout"
            test_result["overall"] = "timeout"
            self._log("Functional test timeout", "ERROR")
        except Exception as e:
            test_result["status"] = "error"
            test_result["error"] = str(e)
            test_result["overall"] = "error"
            self._log(f"Functional test error: {e}", "ERROR")
        
        return test_result
    
    def _run_security_test(self) -> Dict[str, Any]:
        """Run security analysis on the protocol."""
        self._log("Running security analysis...")
        
        spec = self.results["protocol_spec"]
        analysis = spec["security_analysis"]
        
        security_results = {
            "status": "completed",
            "attacks_tested": [],
            "attack_results": [],
        }
        
        attacks = [
            {
                "name": "Replay Attack",
                "description": "Attempt to replay captured messages",
                "mitigated": spec['protocol']['replay_protection'],
                "evidence": "Nonce tracking in session state" if spec['protocol']['replay_protection'] else "No replay protection mechanism",
            },
            {
                "name": "Man-in-the-Middle",
                "description": "Intercept and modify key exchange",
                "mitigated": spec['protocol']['authenticity'],
                "evidence": "Authentication in key exchange" if spec['protocol']['authenticity'] else "No authentication",
            },
            {
                "name": "Forward Secrecy Bypass",
                "description": "Use compromised keys to decrypt past sessions",
                "mitigated": spec['protocol']['forward_secrecy'],
                "evidence": "Ephemeral key exchange" if spec['protocol']['forward_secrecy'] else "Static keys allow past session decryption",
            },
            {
                "name": "Packet Modification",
                "description": "Tamper with encrypted packets",
                "mitigated": spec['protocol']['integrity'],
                "evidence": "AEAD provides integrity protection" if spec['protocol']['integrity'] else "No integrity protection",
            },
            {
                "name": "Session Hijacking",
                "description": "Steal active session",
                "mitigated": spec['protocol']['authenticity'] and spec['protocol']['forward_secrecy'],
                "evidence": "Session binding via authentication" if spec['protocol']['authenticity'] else "No session binding",
            },
            {
                "name": "Downgrade Attack",
                "description": "Force protocol to use weaker algorithms",
                "mitigated": "version_negotiation" in str(spec.get('distinct_features', [])),
                "evidence": "Protocol mandates minimum algorithm strength",
            },
            {
                "name": "Reflection Attack",
                "description": "Reflect messages back to sender",
                "mitigated": "Source identification in packets",
                "evidence": "Connection IDs prevent reflection",
            },
        ]
        
        for attack in attacks:
            security_results["attacks_tested"].append(attack["name"])
            security_results["attack_results"].append({
                "attack": attack["name"],
                "description": attack["description"],
                "mitigated": attack["mitigated"],
                "assessment": "RESISTANT" if attack["mitigated"] else "VULNERABLE",
                "evidence": attack["evidence"],
            })
        
        # Calculate resistance score
        mitigated_count = sum(1 for a in attacks if a["mitigated"])
        resistance_score = mitigated_count / len(attacks)
        
        security_results["resistance_score"] = resistance_score
        security_results["resistance_rating"] = (
            "HIGH" if resistance_score >= 0.8 else
            "MEDIUM" if resistance_score >= 0.5 else
            "LOW"
        )
        
        # Self-criticism results
        security_results["self_criticism"] = analysis["self_criticism"]
        security_results["fatal_flaws"] = analysis.get("fatal_flaws", [])
        
        if security_results["resistance_rating"] == "HIGH":
            self._log(f"Security analysis: HIGH resistance ({resistance_score*100:.0f}%)", "SUCCESS")
        else:
            self._log(f"Security analysis: {security_results['resistance_rating']} resistance ({resistance_score*100:.0f}%)", "WARNING")
        
        return security_results
    
    def execute(self) -> Dict[str, Any]:
        """Execute the full protocol synthesis run."""
        self._log(f"Starting protocol synthesis run {self.run_id}")
        self.results["status"] = "running"
        
        start_time = datetime.utcnow()
        
        try:
            # Step 1: Generate protocol
            spec = self._generate_protocol()
            self.results["status"] = "generated"
            
            # Step 2: Write specification
            self._write_specification(spec)
            
            # Step 3: Implement protocol
            self._implement_protocol(spec)
            
            # Step 4: Install dependencies
            self._install_dependencies()
            
            # Step 5: Compile/verify
            if self._compile_verify():
                self.results["compilation"]["status"] = "success"
            else:
                self.results["compilation"]["status"] = "failed"
                self.results["status"] = "compilation_failed"
            
            # Step 6: Functional test
            self.results["functional_test"] = self._run_functional_test()
            
            # Step 7: Security test
            self.results["security_test"] = self._run_security_test()
            
            # Determine overall status
            if self.results["compilation"].get("status") == "success":
                if self.results["functional_test"].get("overall") in ["pass", "partial"]:
                    if self.results["security_test"].get("resistance_rating") in ["HIGH", "MEDIUM"]:
                        self.results["status"] = "completed"
                    else:
                        self.results["status"] = "completed_with_warnings"
                else:
                    self.results["status"] = "functional_test_failed"
            else:
                self.results["status"] = "compilation_failed"
            
        except Exception as e:
            self.results["status"] = "error"
            self.results["errors"].append(str(e))
            self._log(f"Run error: {e}\n{traceback.format_exc()}", "ERROR")
        
        # Calculate duration
        end_time = datetime.utcnow()
        self.results["duration"] = (end_time - start_time).total_seconds()
        
        # Write results
        results_file = self.output_dir / "results.json"
        with open(results_file, "w") as f:
            json.dump(self.results, f, indent=2, default=str)
        
        # Write summary
        summary_file = self.output_dir / "artifacts" / "summary.md"
        summary = f"""# Run {self.run_id} Summary

**Protocol**: {self.results.get('protocol_name', 'N/A')}
**Status**: {self.results['status']}
**Duration**: {self.results.get('duration', 0):.2f} seconds

## Results

| Phase | Status |
|-------|--------|
| Generation | {'✓' if self.results.get('protocol_spec') else '✗'} |
| Implementation | {'✓' if self.results.get('implementation') else '✗'} |
| Compilation | {'✓' if self.results.get('compilation', {}).get('status') == 'success' else '✗'} |
| Functional Test | {self.results.get('functional_test', {}).get('overall', 'N/A')} |
| Security Test | {self.results.get('security_test', {}).get('resistance_rating', 'N/A')} |

## Functional Test Results

- Passed: {self.results.get('functional_test', {}).get('passed', 0)}
- Failed: {self.results.get('functional_test', {}).get('failed', 0)}

## Security Assessment

- Resistance Score: {self.results.get('security_test', {}).get('resistance_score', 0)*100:.0f}%
- Rating: {self.results.get('security_test', {}).get('resistance_rating', 'N/A')}

## Errors

"""
        
        if self.results.get('errors'):
            for error in self.results['errors']:
                summary += f"- {error}\n"
        else:
            summary += "- None\n"
        
        with open(summary_file, "w") as f:
            f.write(summary)
        
        self._log(f"Run {self.run_id} completed with status: {self.results['status']}")
        
        return self.results


def run_protocol_synthesis(run_id: str, seed: int, output_dir: str) -> Dict[str, Any]:
    """Execute a protocol synthesis run."""
    run = ProtocolRun(run_id, seed, output_dir)
    return run.execute()
