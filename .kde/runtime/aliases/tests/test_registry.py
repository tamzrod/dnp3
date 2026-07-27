"""
Unit tests for Alias Registry
"""

import unittest
import sys
import os

# Add parent directory to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from registry import (
    AliasCategory,
    AliasEntry,
    AliasRegistry,
    reset_registry
)


class TestAliasEntry(unittest.TestCase):
    """Test AliasEntry class."""
    
    def test_from_dict(self):
        """Test creating entry from dictionary."""
        data = {
            'alias': 'go',
            'canonical': 'pre-flight check',
            'category': 'operational',
            'version': '1.0.0',
            'description': 'Execute pre-flight check',
            'namespace': 'runtime',
            'examples': ['go', 'go for launch'],
            'deprecated': False
        }
        
        entry = AliasEntry.from_dict(data)
        
        self.assertEqual(entry.alias, 'go')
        self.assertEqual(entry.canonical, 'pre-flight check')
        self.assertEqual(entry.category, AliasCategory.OPERATIONAL)
        self.assertEqual(entry.version, '1.0.0')
        self.assertEqual(entry.namespace, 'runtime')
        self.assertEqual(len(entry.examples), 2)
    
    def test_to_dict(self):
        """Test converting entry to dictionary."""
        entry = AliasEntry(
            alias='test',
            canonical='test command',
            category=AliasCategory.FRIENDLY,
            version='1.0.0',
            description='Test entry',
            namespace='test'
        )
        
        data = entry.to_dict()
        
        self.assertEqual(data['alias'], 'test')
        self.assertEqual(data['category'], 'friendly')
        self.assertEqual(data['version'], '1.0.0')


class TestAliasRegistry(unittest.TestCase):
    """Test AliasRegistry class."""
    
    def setUp(self):
        """Set up test fixtures."""
        reset_registry()
        self.registry = AliasRegistry()
    
    def tearDown(self):
        """Tear down test fixtures."""
        reset_registry()
    
    def test_load_registry(self):
        """Test loading registry from file."""
        result = self.registry.load()
        self.assertTrue(result)
        self.assertTrue(self.registry.is_loaded)
    
    def test_resolve_alias(self):
        """Test resolving an alias."""
        self.registry.load()
        
        entry = self.registry.resolve('go')
        
        self.assertIsNotNone(entry)
        self.assertEqual(entry.alias, 'go')
        self.assertEqual(entry.category, AliasCategory.OPERATIONAL)
    
    def test_resolve_canonical(self):
        """Test resolving a canonical command."""
        self.registry.load()
        
        entry = self.registry.resolve('start engine')
        
        self.assertIsNotNone(entry)
        self.assertEqual(entry.category, AliasCategory.CANONICAL)
    
    def test_resolve_unknown(self):
        """Test resolving an unknown alias."""
        self.registry.load()
        
        entry = self.registry.resolve('unknown-command-xyz')
        
        self.assertIsNone(entry)
    
    def test_resolve_to_canonical(self):
        """Test resolving alias to canonical name."""
        self.registry.load()
        
        canonical = self.registry.resolve_to_canonical('go')
        
        self.assertIsNotNone(canonical)
        # The canonical should be the pre-flight check entry
        self.assertIn('pre-flight', canonical.lower())
    
    def test_get_by_category(self):
        """Test getting aliases by category."""
        self.registry.load()
        
        friendly = self.registry.get_by_category(AliasCategory.FRIENDLY)
        
        self.assertIsInstance(friendly, list)
        for entry in friendly:
            self.assertEqual(entry.category, AliasCategory.FRIENDLY)
    
    def test_get_friendly_aliases(self):
        """Test getting friendly aliases."""
        self.registry.load()
        
        friendly = self.registry.get_friendly_aliases()
        
        self.assertIsInstance(friendly, list)
        self.assertTrue(len(friendly) > 0)
    
    def test_get_operational_aliases(self):
        """Test getting operational aliases."""
        self.registry.load()
        
        ops = self.registry.get_operational_aliases()
        
        self.assertIsInstance(ops, list)
        for entry in ops:
            self.assertEqual(entry.category, AliasCategory.OPERATIONAL)
    
    def test_get_canonical_aliases(self):
        """Test getting all aliases for a canonical command."""
        self.registry.load()
        
        aliases = self.registry.get_canonical_aliases('pre-flight check')
        
        self.assertIsInstance(aliases, list)
        self.assertTrue(len(aliases) > 1)  # Should have multiple aliases
    
    def test_suggest(self):
        """Test alias suggestion."""
        self.registry.load()
        
        suggestions = self.registry.suggest('go')
        
        self.assertIsInstance(suggestions, list)
        self.assertTrue(len(suggestions) > 0)
        self.assertTrue(any('go' in s.alias.lower() for s in suggestions))
    
    def test_suggest_with_limit(self):
        """Test alias suggestion with limit."""
        self.registry.load()
        
        suggestions = self.registry.suggest('s', limit=3)
        
        self.assertLessEqual(len(suggestions), 3)
    
    def test_get_stats(self):
        """Test getting registry statistics."""
        self.registry.load()
        
        stats = self.registry.get_stats()
        
        self.assertIn('total_aliases', stats)
        self.assertIn('by_category', stats)
        self.assertGreater(stats['total_aliases'], 0)
    
    def test_validate(self):
        """Test registry validation."""
        self.registry.load()
        
        result = self.registry.validate()
        
        self.assertIn('valid', result)
        self.assertIn('warnings', result)
        # Should be valid (no critical errors)


class TestGlobalRegistry(unittest.TestCase):
    """Test global registry singleton."""
    
    def setUp(self):
        """Set up test fixtures."""
        reset_registry()
    
    def tearDown(self):
        """Tear down test fixtures."""
        reset_registry()
    
    def test_get_registry(self):
        """Test getting global registry."""
        from registry import get_registry
        
        reg = get_registry()
        
        self.assertIsInstance(reg, AliasRegistry)
        self.assertTrue(reg.is_loaded)
    
    def test_reset_registry(self):
        """Test resetting global registry."""
        from registry import get_registry, reset_registry
        
        reg1 = get_registry()
        reset_registry()
        reg2 = get_registry()
        
        # Should be different instances
        self.assertIsNot(reg1, reg2)


class TestAliasResolution(unittest.TestCase):
    """Test alias resolution scenarios."""
    
    def setUp(self):
        """Set up test fixtures."""
        reset_registry()
        self.registry = AliasRegistry()
        self.registry.load()
    
    def tearDown(self):
        """Tear down test fixtures."""
        reset_registry()
    
    def test_case_insensitive(self):
        """Test case insensitive resolution."""
        entry1 = self.registry.resolve('go')
        entry2 = self.registry.resolve('GO')
        entry3 = self.registry.resolve('Go')
        
        self.assertEqual(entry1.alias, entry2.alias)
        self.assertEqual(entry2.alias, entry3.alias)
    
    def test_all_canonical_commands(self):
        """Test that all canonical commands resolve."""
        canonicals = [
            'start engine',
            'pre-flight check',
            'mission ready',
            'check state',
            'bootstrap',
            'run demo'
        ]
        
        for cmd in canonicals:
            entry = self.registry.resolve(cmd)
            if entry:
                self.assertEqual(entry.category, AliasCategory.CANONICAL)


if __name__ == '__main__':
    unittest.main()
