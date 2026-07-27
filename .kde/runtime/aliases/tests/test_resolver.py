"""
Unit tests for Command Resolver
"""

import unittest
import sys
import os

# Add parent directory to path
test_dir = os.path.dirname(__file__)
parent_dir = os.path.dirname(test_dir)
sys.path.insert(0, parent_dir)

from runtime.aliases.resolver import (
    CommandType,
    ResolvedCommand,
    CommandParser,
    RuntimeCommandHandler,
    create_parser,
    create_handler
)
from runtime.aliases.registry import reset_registry


class TestResolvedCommand(unittest.TestCase):
    """Test ResolvedCommand dataclass."""
    
    def test_create_canonical(self):
        """Test creating a canonical resolved command."""
        resolved = ResolvedCommand(
            command_type=CommandType.CANONICAL,
            original_input='start engine',
            resolved_command='start engine'
        )
        
        self.assertEqual(resolved.command_type, CommandType.CANONICAL)
        self.assertFalse(resolved.deprecated)
    
    def test_create_alias(self):
        """Test creating an alias resolved command."""
        resolved = ResolvedCommand(
            command_type=CommandType.ALIAS,
            original_input='go',
            resolved_command='pre-flight check',
            deprecated=False
        )
        
        self.assertEqual(resolved.command_type, CommandType.ALIAS)
        self.assertFalse(resolved.deprecated)
    
    def test_create_deprecated(self):
        """Test creating a deprecated resolved command."""
        resolved = ResolvedCommand(
            command_type=CommandType.ALIAS,
            original_input='old-cmd',
            resolved_command='new-cmd',
            deprecated=True,
            deprecation_warning="WARNING: 'old-cmd' is deprecated"
        )
        
        self.assertTrue(resolved.deprecated)
        self.assertIsNotNone(resolved.deprecation_warning)


class TestCommandParser(unittest.TestCase):
    """Test CommandParser class."""
    
    def setUp(self):
        """Set up test fixtures."""
        reset_registry()
        self.parser = create_parser()
    
    def tearDown(self):
        """Tear down test fixtures."""
        reset_registry()
    
    def test_parse_canonical(self):
        """Test parsing canonical command."""
        resolved = self.parser.parse('start engine')
        
        self.assertEqual(resolved.command_type, CommandType.CANONICAL)
        self.assertEqual(resolved.resolved_command, 'start engine')
    
    def test_parse_alias(self):
        """Test parsing alias."""
        resolved = self.parser.parse('go')
        
        self.assertEqual(resolved.command_type, CommandType.ALIAS)
        self.assertIn('pre-flight', resolved.resolved_command.lower())
    
    def test_parse_case_insensitive(self):
        """Test case insensitive parsing."""
        resolved1 = self.parser.parse('GO')
        resolved2 = self.parser.parse('Go')
        resolved3 = self.parser.parse('go')
        
        self.assertEqual(resolved1.command_type, resolved2.command_type)
        self.assertEqual(resolved2.command_type, resolved3.command_type)
    
    def test_parse_unknown(self):
        """Test parsing unknown command."""
        resolved = self.parser.parse('unknown-xyz-123')
        
        self.assertEqual(resolved.command_type, CommandType.UNKNOWN)
    
    def test_parse_with_whitespace(self):
        """Test parsing with extra whitespace."""
        resolved = self.parser.parse('  go  ')
        
        self.assertEqual(resolved.command_type, CommandType.ALIAS)
    
    def test_reserved_words(self):
        """Test reserved words are not treated as aliases."""
        for word in ['help', 'exit', 'quit']:
            resolved = self.parser.parse(word)
            self.assertEqual(resolved.command_type, CommandType.CANONICAL)
    
    def test_suggest(self):
        """Test suggestion."""
        suggestions = self.parser.suggest('go')
        
        self.assertIsInstance(suggestions, list)
        self.assertTrue(len(suggestions) > 0)
        self.assertTrue(any('go' in s.lower() for s in suggestions))
    
    def test_suggest_with_limit(self):
        """Test suggestion with limit."""
        suggestions = self.parser.suggest('s', limit=3)
        
        self.assertLessEqual(len(suggestions), 3)


class TestRuntimeCommandHandler(unittest.TestCase):
    """Test RuntimeCommandHandler class."""
    
    def setUp(self):
        """Set up test fixtures."""
        reset_registry()
        self.handler = create_handler()
    
    def tearDown(self):
        """Tear down test fixtures."""
        reset_registry()
    
    def test_handle_canonical(self):
        """Test handling canonical command."""
        success, message, warning = self.handler.handle('start engine')
        
        self.assertTrue(success)
        self.assertIn('resolved', message.lower())
        self.assertIsNone(warning)
    
    def test_handle_alias(self):
        """Test handling alias."""
        success, message, warning = self.handler.handle('go')
        
        self.assertTrue(success)
        self.assertIn('resolved', message.lower())
        self.assertIsNone(warning)
    
    def test_handle_unknown(self):
        """Test handling unknown command."""
        success, message, warning = self.handler.handle('unknown-xyz')
        
        self.assertFalse(success)
        self.assertIn('unknown', message.lower())
    
    def test_handle_unknown_no_suggestion(self):
        """Test handling unknown with no suggestions."""
        success, message, warning = self.handler.handle('zzz-not-exist')
        
        self.assertFalse(success)
        self.assertIn('unknown', message.lower())
    
    def test_list_commands(self):
        """Test listing commands."""
        commands = self.handler.list_commands()
        
        self.assertIsInstance(commands, list)
        self.assertTrue(len(commands) > 0)
    
    def test_list_commands_by_category(self):
        """Test listing commands by category."""
        commands = self.handler.list_commands(category='operational')
        
        self.assertIsInstance(commands, list)


class TestCreateFunctions(unittest.TestCase):
    """Test factory functions."""
    
    def setUp(self):
        """Set up test fixtures."""
        reset_registry()
    
    def tearDown(self):
        """Tear down test fixtures."""
        reset_registry()
    
    def test_create_parser(self):
        """Test creating parser."""
        parser = create_parser()
        
        self.assertIsInstance(parser, CommandParser)
    
    def test_create_handler(self):
        """Test creating handler."""
        handler = create_handler()
        
        self.assertIsInstance(handler, RuntimeCommandHandler)


if __name__ == '__main__':
    unittest.main()
