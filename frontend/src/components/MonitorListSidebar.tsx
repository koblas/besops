import { Input, Button, Empty, Spin } from 'antd';
import { PlusOutlined, SearchOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useMemo, useState } from 'react';
import { useMonitors } from '../hooks/useMonitors';
import { useUptimes } from '../hooks/useUptimes';
import type { Monitor } from '../hooks/useMonitors';
import { MonitorListItem } from './MonitorListItem';

export function MonitorListSidebar() {
  const navigate = useNavigate();
  const { id: activeId } = useParams<{ id: string }>();
  const [search, setSearch] = useState('');
  const { data: monitors, isLoading, isError } = useMonitors();
  const { data: uptimes = {} } = useUptimes();

  const { roots, childrenMap, matchesSearch } = useMemo(() => {
    if (!monitors) return { roots: [], childrenMap: new Map<string, Monitor[]>(), matchesSearch: new Set<string>() };

    const q = search.toLowerCase().trim();
    const map = new Map<string, Monitor[]>();
    const matched = new Set<string>();

    for (const m of monitors) {
      const parentId = m.parentId ?? '__root__';
      const list = map.get(parentId) ?? [];
      list.push(m);
      map.set(parentId, list);

      if (!q || m.name.toLowerCase().includes(q)) {
        matched.add(m.id);
      }
    }

    // If searching, also include parents of matched monitors so the tree is visible
    if (q) {
      for (const m of monitors) {
        if (matched.has(m.id) && m.parentId) {
          let parentId: string | undefined = m.parentId;
          while (parentId) {
            matched.add(parentId);
            const parent = monitors.find(p => p.id === parentId);
            parentId = parent?.parentId;
          }
        }
      }
    }

    const rootMonitors = (map.get('__root__') ?? []).filter(m => matched.has(m.id));

    return { roots: rootMonitors, childrenMap: map, matchesSearch: matched };
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
        ) : roots.length === 0 ? (
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
          roots.map(monitor => (
            <MonitorListItem
              key={monitor.id}
              monitor={monitor}
              active={monitor.id === activeId}
              activeId={activeId}
              depth={0}
              childrenMap={childrenMap}
              matchesSearch={matchesSearch}
              uptimes={uptimes}
              onClick={id => navigate(`/dashboard/${id}`)}
            />
          ))
        )}
      </div>
    </div>
  );
}
