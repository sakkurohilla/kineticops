import React, { useEffect, useState } from 'react';
import Card from '../common/Card';
import Input from '../common/Input';
import { useId } from 'react';
import { AlertRuleRequest } from '../../hooks/useAlerts';

interface Props {
  rules: any[];
  onCreate: (r: AlertRuleRequest) => Promise<any>;
  loading?: boolean;
}

const RuleEditor: React.FC<Props> = ({ rules, onCreate }) => {
  const [form, setForm] = useState<AlertRuleRequest>({
    metric_name: '',
    operator: '>',
    threshold: 0,
    window: 5,
    frequency: 1,
    notification_webhook: '',
    escalation_policy: '',
  });

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setError(null);
  }, [form]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      await onCreate(form);
      // reset minimal fields
      setForm((f) => ({ ...f, metric_name: '', threshold: 0 }));
    } catch (err: any) {
      setError(err?.message || 'Failed to create rule');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <h2 className="text-lg font-semibold mb-4">Alert Rules</h2>

      <Card className="mb-4">
        <form onSubmit={handleSubmit} className="space-y-3">
          <div>
            <Input
              label="Metric Name"
              value={form.metric_name}
              onChange={(e) => setForm({ ...form, metric_name: e.target.value })}
              placeholder="metric.name"
            />
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="block text-sm font-medium text-gray-700">Operator</label>
              <select value={form.operator} onChange={(e) => setForm({ ...form, operator: e.target.value })} className="mt-1 block w-full border rounded p-2">
                <option value=">">&gt;</option>
                <option value="<">&lt;</option>
                <option value=">=">&gt;=</option>
                <option value="<=">&lt;=</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Threshold</label>
              <Input type="number" value={form.threshold} onChange={(e) => setForm({ ...form, threshold: Number(e.target.value) })} />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Window (min)</label>
              <Input type="number" value={form.window} onChange={(e) => setForm({ ...form, window: Number(e.target.value) })} />
            </div>
          </div>

          <div>
            <Input
              label="Notification Webhook"
              value={form.notification_webhook}
              onChange={(e) => setForm({ ...form, notification_webhook: e.target.value })}
              placeholder="https://hooks.example.com/alert"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">Escalation Policy (JSON)</label>
            <textarea id={`escalation-${useId()}`} name={`escalation-${useId()}`} value={form.escalation_policy} onChange={(e) => setForm({ ...form, escalation_policy: e.target.value })} rows={3} className="mt-1 block w-full border rounded p-2" />
          </div>

          {error && <div className="text-sm text-red-600">{error}</div>}

          <div className="flex items-center gap-2">
            <button type="submit" disabled={submitting} className="px-4 py-2 rounded bg-blue-600 text-white">{submitting ? 'Saving...' : 'Create Rule'}</button>
            <button type="button" onClick={() => setForm({ metric_name: '', operator: '>', threshold: 0, window: 5, frequency: 1, notification_webhook: '', escalation_policy: '' })} className="px-3 py-2 rounded bg-gray-100">Reset</button>
          </div>
        </form>
      </Card>

      <Card>
        <h4 className="font-medium mb-2">Existing Rules</h4>
        {rules.length === 0 ? (
          <div className="text-sm text-gray-500">No alert rules configured.</div>
        ) : (
          <ul className="space-y-2">
            {rules.map((r: any) => (
              <li key={r.id} className="border rounded p-3">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-sm font-medium">{r.metric_name} {r.operator} {r.threshold}</div>
                    <div className="text-xs text-gray-500">Window: {r.window}m • Freq: {r.frequency}</div>
                  </div>
                  <div className="text-xs text-gray-400">Created: {new Date(r.created_at).toLocaleString()}</div>
                </div>
              </li>
            ))}
          </ul>
        )}
      </Card>
    </div>
  );
};

export default RuleEditor;
