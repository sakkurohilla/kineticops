import React from 'react';
import MainLayout from '../../components/layout/MainLayout';
import useAlerts from '../../hooks/useAlerts';
import AlertList from '../../components/alerts/AlertList';
import RuleEditor from '../../components/alerts/RuleEditor';

const AlertsPage: React.FC = () => {
  const { alerts, rules, loading, error, createRule, silenceAlert } = useAlerts();

  return (
    <MainLayout>
      <div className="p-6">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-semibold">Alerts</h1>
        </div>

        <div className="grid grid-cols-12 gap-6">
          <div className="col-span-7">
            <AlertList alerts={alerts} onSilence={async (id) => {
              await silenceAlert(id);
            }} />
          </div>

          <div className="col-span-5">
            <RuleEditor rules={rules} onCreate={createRule} loading={loading} />
          </div>
        </div>

        {error && <div className="mt-4 text-sm text-red-600">{error}</div>}
      </div>
    </MainLayout>
  );
};

export default AlertsPage;
