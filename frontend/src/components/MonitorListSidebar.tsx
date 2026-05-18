import { Input, Button, Empty, Spin } from 'antd';
import { PlusOutlined, SearchOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useMemo, useState } from 'react';
import { useMonitors } from '../hooks/useMonitors';
import { useUptimes } from '../hooks/useUptimes';
import { MonitorListItem } from './MonitorListItem';

export function MonitorListSidebar() {
  const navigate = useNavigate();
  const { id: activeId } = useParams<{ id: string }>();
  const [search, setSearch] = useState('');
  const { data: monitors, isLoading, isError } = useMonitors();
  const { data: uptimes = {} } = useUptimes();

  const filtered = useMemo(() => {
    if (!monitors) return [];

    const q = search.toLowerCase().trim();
    const list = q
      ? monitors.filter(m => m.name.toLowerCase().includes(q))
      : [...monitors];

    list.sort((a, b) => {
      const aGroup = a.type === 'group' ? 0 : 1;
      const bGroup = b.type === 'group' ? 0 : 1;
      if (aGroup !== bGroup) return aGroup - bGroup;
      return a.name.localeCompare(b.name);
    });

    return list;
  }, [monitors, search]);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', padding: '12px 8px' }}>
      <div style={{ marginBottom: 16 }}>
        <Input
          prefix={<SearchOutlined />}
          placeholder="Search..."
          value={search}
          onChange={e => setSearch(e.target.value)}
          allowClear
          size="small"
        />
      </div>
      <div style={{ marginBottom: 16 }}>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate('/add')}
          size="small"
          block
        >
          Add Monitor
        </Button>
      </div>
      <div style={{ flex: 1, overflow: 'auto' }}>
        {isError ? (
          <Empty description="Failed to load monitors" style={{ marginTop: 48 }} />
        ) : isLoading ? (
          <div style={{ textAlign: 'center', marginTop: 48 }}>
            <Spin />
          </div>
        ) : filtered.length === 0 ? (
          <Empty
            description={search ? 'No matches found' : 'No monitors yet'}
            style={{ marginTop: 48 }}
          >
            {!search && (
              <Button type="primary" size="small" icon={<PlusOutlined />} onClick={() => navigate('/add')}>
                Add Monitor
              </Button>
            )}
          </Empty>
        ) : (
          filtered.map(monitor => (
            <MonitorListItem
              key={monitor.id}
              monitor={monitor}
              active={monitor.id === activeId}
              uptimes={uptimes}
              onClick={id => navigate(`/dashboard/${id}`)}
            />
          ))
        )}
      </div>
    </div>
  );
}
