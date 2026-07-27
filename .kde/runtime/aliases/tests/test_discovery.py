"""
Unit tests for Discovery API
"""

import unittest
import sys
import os

# Add parent directory to path
test_dir = os.path.dirname(__file__)
parent_dir = os.path.dirname(test_dir)
sys.path.insert(0, parent_dir)

from runtime.aliases.discovery import (
    DiscoveryAPI,
    DiscoveryEndpoint,
    create_api
)
from runtime.aliases.registry import reset_registry


class TestDiscoveryAPI(unittest.TestCase):
    """Test DiscoveryAPI class."""
    
    def setUp(self):
        """Set up test fixtures."""
        reset_registry()
        self.api = create_api()
    
    def tearDown(self):
        """Tear down test fixtures."""
        reset_registry()
    
    def test_list_all(self):
        """Test listing all aliases."""
        response = self.api.list_all()
        
        self.assertTrue(response.success)
        self.assertEqual(response.endpoint, 'list_all')
        self.assertIn('aliases', response.data)
        self.assertGreater(response.data['count'], 0)
    
    def test_list_by_category(self):
        """Test listing by category."""
        response = self.api.list_by_category('friendly')
        
        self.assertTrue(response.success)
        self.assertEqual(response.data['category'], 'friendly')
        for alias in response.data['aliases']:
            self.assertEqual(alias['category'], 'friendly')
    
    def test_list_by_category_invalid(self):
        """Test listing with invalid category."""
        response = self.api.list_by_category('invalid_category')
        
        self.assertFalse(response.success)
        self.assertIsNotNone(response.error)
    
    def test_resolve_known_alias(self):
        """Test resolving a known alias."""
        response = self.api.resolve('go')
        
        self.assertTrue(response.success)
        self.assertEqual(response.data['alias'], 'go')
        self.assertIn('canonical', response.data)
    
    def test_resolve_unknown_alias(self):
        """Test resolving an unknown alias."""
        response = self.api.resolve('unknown-xyz-123')
        
        self.assertFalse(response.success)
        self.assertIsNotNone(response.error)
    
    def test_get_canonical(self):
        """Test getting canonical with aliases."""
        response = self.api.get_canonical('pre-flight check')
        
        self.assertTrue(response.success)
        self.assertIsNotNone(response.data['canonical'])
        self.assertGreater(response.data['total_count'], 1)
    
    def test_get_canonical_unknown(self):
        """Test getting unknown canonical."""
        response = self.api.get_canonical('unknown-command')
        
        self.assertFalse(response.success)
    
    def test_suggest(self):
        """Test suggesting aliases."""
        response = self.api.suggest('go')
        
        self.assertTrue(response.success)
        self.assertGreater(response.data['count'], 0)
        for suggestion in response.data['suggestions']:
            self.assertTrue(suggestion['alias'].startswith('go'))
    
    def test_suggest_with_limit(self):
        """Test suggestion with limit."""
        response = self.api.suggest('s', limit=3)
        
        self.assertLessEqual(response.data['count'], 3)
    
    def test_suggest_with_category(self):
        """Test suggestion with category filter."""
        response = self.api.suggest('s', category='operational')
        
        self.assertTrue(response.success)
        for suggestion in response.data['suggestions']:
            self.assertEqual(suggestion['category'], 'operational')
    
    def test_get_stats(self):
        """Test getting statistics."""
        response = self.api.get_stats()
        
        self.assertTrue(response.success)
        self.assertIn('total_aliases', response.data)
        self.assertIn('by_category', response.data)
    
    def test_validate(self):
        """Test registry validation."""
        response = self.api.validate()
        
        self.assertIn('valid', response.data)
        self.assertIn('warnings', response.data)


class TestCreateAPI(unittest.TestCase):
    """Test create_api factory function."""
    
    def test_create_api(self):
        """Test creating API instance."""
        api = create_api()
        
        self.assertIsInstance(api, DiscoveryAPI)


if __name__ == '__main__':
    unittest.main()
