#!/bin/bash
# PATCH-001 Migration Checklist
# Run this script to deploy the WorkspaceResolver component

set -e

echo "=========================================="
echo "PATCH-001: Workspace Resolver Migration"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

WORKSPACE_ROOT="/workspace/project/kde"
BACKUP_DIR="/tmp/patch-001-backup-$(date +%Y%m%d-%H%M%S)"

# Step 1: Backup
echo -e "${YELLOW}[1/7]${NC} Creating backup..."
mkdir -p "$BACKUP_DIR"
cp "$WORKSPACE_ROOT/runtime/orchestrator/__init__.py" "$BACKUP_DIR/"
echo "Backup created at: $BACKUP_DIR"
echo ""

# Step 2: Check syntax of new files
echo -e "${YELLOW}[2/7]${NC} Checking syntax of new files..."
python3 -m py_compile "$WORKSPACE_ROOT/runtime/orchestrator/workspace.py" || {
    echo -e "${RED}ERROR: workspace.py has syntax errors${NC}"
    exit 1
}
python3 -m py_compile "$WORKSPACE_ROOT/runtime/orchestrator/types.py" || {
    echo -e "${RED}ERROR: types.py has syntax errors${NC}"
    exit 1
}
echo -e "${GREEN}Syntax check passed${NC}"
echo ""

# Step 3: Add new files
echo -e "${YELLOW}[3/7]${NC} Adding new files..."
if [ ! -f "$WORKSPACE_ROOT/runtime/orchestrator/workspace.py" ]; then
    echo "ERROR: workspace.py not found"
    exit 1
fi
if [ ! -f "$WORKSPACE_ROOT/runtime/orchestrator/types.py" ]; then
    echo "ERROR: types.py not found"
    exit 1
fi
echo -e "${GREEN}New files present${NC}"
echo ""

# Step 4: Modify __init__.py
echo -e "${YELLOW}[4/7]${NC} Modifying __init__.py..."
echo "This step requires manual intervention:"
echo "  - Add 'from .workspace import WorkspaceResolver, WorkspaceInfo' to imports"
echo "  - Add 'self.workspace_resolver = WorkspaceResolver()' to __init__"
echo "  - Add workspace resolution step after classification"
echo "  - Add workspace field to OrchestrationResult"
echo ""
read -p "Have you made the modifications? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Migration cancelled"
    exit 1
fi
echo ""

# Step 5: Run tests
echo -e "${YELLOW}[5/7]${NC} Running validation tests..."
python3 "$WORKSPACE_ROOT/governance/runtime/PATCH-001/VALIDATION-TESTS.py" || {
    echo -e "${RED}ERROR: Tests failed${NC}"
    exit 1
}
echo ""

# Step 6: Verify import
echo -e "${YELLOW}[6/7]${NC} Verifying imports..."
python3 -c "from runtime.orchestrator import WorkspaceResolver, WorkspaceInfo; print('Import OK')" || {
    echo -e "${RED}ERROR: Import failed${NC}"
    exit 1
}
echo -e "${GREEN}Import verification passed${NC}"
echo ""

# Step 7: Summary
echo -e "${YELLOW}[7/7]${NC} Migration complete!"
echo ""
echo "=========================================="
echo "Summary"
echo "=========================================="
echo "Backup location: $BACKUP_DIR"
echo ""
echo "Files added:"
echo "  - runtime/orchestrator/workspace.py"
echo "  - runtime/orchestrator/types.py"
echo ""
echo "Files modified:"
echo "  - runtime/orchestrator/__init__.py"
echo ""
echo "Rollback command:"
echo "  cp $BACKUP_DIR/__init__.py $WORKSPACE_ROOT/runtime/orchestrator/__init__.py"
echo "  rm $WORKSPACE_ROOT/runtime/orchestrator/workspace.py"
echo "  rm $WORKSPACE_ROOT/runtime/orchestrator/types.py"
echo ""
echo "=========================================="
echo -e "${GREEN}PATCH-001 migration completed successfully${NC}"
echo "=========================================="
