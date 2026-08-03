"""pytest conftest — add the agentic-os directory to sys.path."""
import sys
import os
sys.path.insert(0, os.path.dirname(__file__))
