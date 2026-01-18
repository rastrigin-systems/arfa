'use client';

import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Plus, Trash2 } from 'lucide-react';
import type { ToolPolicy, UpdateToolPolicyRequest } from '@/lib/types';

type EditPolicyModalProps = {
  isOpen: boolean;
  onClose: () => void;
  onEdit: (data: UpdateToolPolicyRequest) => void;
  policy: ToolPolicy;
  isLoading?: boolean;
};

type Condition = {
  param_path: string;
  operator: 'contains' | 'matches' | 'equals';
  value: string;
};

// Common tool name suggestions
const TOOL_SUGGESTIONS = [
  { value: 'Bash', description: 'Shell command execution' },
  { value: 'Read', description: 'File reading operations' },
  { value: 'Write', description: 'File writing operations' },
  { value: 'Edit', description: 'File editing operations' },
  { value: 'mcp__*', description: 'All MCP tools' },
  { value: '*', description: 'All tools (use with caution)' },
];

// Common param paths for different tools
const PARAM_SUGGESTIONS: Record<string, string[]> = {
  Bash: ['command'],
  Read: ['file_path'],
  Write: ['file_path', 'content'],
  Edit: ['file_path', 'old_string', 'new_string'],
  default: ['file_path', 'command', 'content'],
};

// Parse existing conditions from policy
// Format: { "file_path": { "contains": ".env" } }
function parseConditions(conditions: object | null | undefined): Condition[] {
  if (!conditions) return [];

  const result: Condition[] = [];
  const condObj = conditions as Record<string, Record<string, string>>;

  for (const [paramPath, operators] of Object.entries(condObj)) {
    if (typeof operators === 'object' && operators !== null) {
      for (const [operator, value] of Object.entries(operators)) {
        result.push({
          param_path: paramPath,
          operator: operator as Condition['operator'],
          value: String(value),
        });
      }
    }
  }

  return result;
}

// Build conditions object for API
// Format: { "file_path": { "contains": ".env" } }
function buildConditions(conditions: Condition[]): Record<string, Record<string, string>> | null {
  const validConditions = conditions.filter((c) => c.param_path && c.value);
  if (validConditions.length === 0) return null;

  const result: Record<string, Record<string, string>> = {};
  for (const cond of validConditions) {
    if (!result[cond.param_path]) {
      result[cond.param_path] = {};
    }
    result[cond.param_path][cond.operator] = cond.value;
  }
  return result;
}

export function EditPolicyModal({
  isOpen,
  onClose,
  onEdit,
  policy,
  isLoading,
}: EditPolicyModalProps) {
  const [toolName, setToolName] = useState(policy.tool_name);
  const [action, setAction] = useState<'deny' | 'audit'>(policy.action);
  const [reason, setReason] = useState(policy.reason || '');
  const [conditions, setConditions] = useState<Condition[]>(() =>
    parseConditions(policy.conditions)
  );
  const [errors, setErrors] = useState<Record<string, string>>({});

  const getParamSuggestions = () => {
    return PARAM_SUGGESTIONS[toolName] || PARAM_SUGGESTIONS.default;
  };

  const addCondition = () => {
    setConditions([...conditions, { param_path: '', operator: 'contains', value: '' }]);
  };

  const updateCondition = (index: number, field: keyof Condition, value: string) => {
    const updated = [...conditions];
    updated[index] = { ...updated[index], [field]: value };
    setConditions(updated);
  };

  const removeCondition = (index: number) => {
    setConditions(conditions.filter((_, i) => i !== index));
  };

  // Reset form when policy changes
  useEffect(() => {
    setToolName(policy.tool_name);
    setAction(policy.action);
    setReason(policy.reason || '');
    setConditions(parseConditions(policy.conditions));
    setErrors({});
  }, [policy]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrors({});

    // Validation
    const newErrors: Record<string, string> = {};
    if (!toolName.trim()) {
      newErrors.toolName = 'Tool name is required';
    }

    // Validate conditions
    conditions.forEach((cond, index) => {
      if (cond.param_path && !cond.value) {
        newErrors[`condition_${index}`] = 'Value is required when param is set';
      }
    });

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    await onEdit({
      tool_name: toolName.trim(),
      action,
      reason: reason.trim() || null,
      conditions: buildConditions(conditions) as UpdateToolPolicyRequest['conditions'],
    });
  };

  const handleClose = () => {
    setErrors({});
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && handleClose()}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Policy</DialogTitle>
          <DialogDescription>
            Update the tool policy settings
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Tool Name */}
          <div className="space-y-2">
            <Label htmlFor="toolName">
              Tool Name <span className="text-red-500">*</span>
            </Label>
            <Input
              id="toolName"
              placeholder="e.g., Bash, Read, mcp__playwright__*"
              value={toolName}
              onChange={(e) => setToolName(e.target.value)}
              className={errors.toolName ? 'border-red-500' : ''}
              list="tool-suggestions-edit"
            />
            <datalist id="tool-suggestions-edit">
              {TOOL_SUGGESTIONS.map((tool) => (
                <option key={tool.value} value={tool.value}>
                  {tool.description}
                </option>
              ))}
            </datalist>
            {errors.toolName && <p className="text-sm text-red-500">{errors.toolName}</p>}
            <p className="text-xs text-muted-foreground">
              Use * as wildcard (e.g., mcp__* matches all MCP tools)
            </p>
          </div>

          {/* Action */}
          <div className="space-y-2">
            <Label htmlFor="action">Action</Label>
            <Select value={action} onValueChange={(v) => setAction(v as 'deny' | 'audit')}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="deny">
                  <span className="flex items-center gap-2">
                    <span className="h-2 w-2 rounded-full bg-red-500" />
                    <span>Deny</span>
                    <span className="text-muted-foreground">- Block tool execution</span>
                  </span>
                </SelectItem>
                <SelectItem value="audit">
                  <span className="flex items-center gap-2">
                    <span className="h-2 w-2 rounded-full bg-amber-500" />
                    <span>Audit</span>
                    <span className="text-muted-foreground">- Log but allow</span>
                  </span>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Scope (read-only) */}
          <div className="space-y-2">
            <Label>Scope</Label>
            <div className="rounded-md border bg-muted/50 px-3 py-2 text-sm">
              {policy.scope === 'organization' && 'Organization-wide'}
              {policy.scope === 'team' && 'Team-level'}
              {policy.scope === 'employee' && 'Employee-specific'}
            </div>
            <p className="text-xs text-muted-foreground">
              Scope cannot be changed after creation
            </p>
          </div>

          {/* Conditions */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <Label>Conditions (Optional)</Label>
              <Button type="button" variant="outline" size="sm" onClick={addCondition}>
                <Plus className="h-4 w-4 mr-1" />
                Add Condition
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              Add conditions to match specific parameter values (e.g., block writes to .env files)
            </p>

            {conditions.length > 0 && (
              <div className="space-y-3 border rounded-lg p-3 bg-muted/30">
                {conditions.map((condition, index) => (
                  <div key={index} className="space-y-2">
                    {index > 0 && (
                      <div className="text-xs text-muted-foreground text-center py-1">OR</div>
                    )}
                    <div className="flex gap-2 items-start">
                      <div className="flex-1 space-y-1">
                        <Input
                          placeholder="Parameter (e.g., file_path)"
                          value={condition.param_path}
                          onChange={(e) => updateCondition(index, 'param_path', e.target.value)}
                          list={`param-suggestions-edit-${index}`}
                          className="text-sm"
                        />
                        <datalist id={`param-suggestions-edit-${index}`}>
                          {getParamSuggestions().map((param) => (
                            <option key={param} value={param} />
                          ))}
                        </datalist>
                      </div>
                      <Select
                        value={condition.operator}
                        onValueChange={(v) =>
                          updateCondition(index, 'operator', v as Condition['operator'])
                        }
                      >
                        <SelectTrigger className="w-[120px]">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="contains">contains</SelectItem>
                          <SelectItem value="matches">matches (regex)</SelectItem>
                          <SelectItem value="equals">equals</SelectItem>
                        </SelectContent>
                      </Select>
                      <div className="flex-1">
                        <Input
                          placeholder="Value (e.g., .env)"
                          value={condition.value}
                          onChange={(e) => updateCondition(index, 'value', e.target.value)}
                          className="text-sm"
                        />
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => removeCondition(index)}
                        className="h-9 w-9 text-muted-foreground hover:text-destructive"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                    {errors[`condition_${index}`] && (
                      <p className="text-sm text-red-500">{errors[`condition_${index}`]}</p>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Reason */}
          <div className="space-y-2">
            <Label htmlFor="reason">Reason (Optional)</Label>
            <Textarea
              id="reason"
              placeholder="Describe why this policy exists..."
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              rows={3}
            />
            <p className="text-xs text-muted-foreground">
              This message will be shown to agents when the policy is triggered
            </p>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose} disabled={isLoading}>
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
