import { useEffect } from 'react';
import { useMonitors } from '../hooks/useMonitors';

export function FaviconBadge() {
  const { data: monitors } = useMonitors();

  useEffect(() => {
    if (!monitors) return;

    const downCount = monitors.filter(m => {
      // We don't have real-time status on the Monitor object from the list endpoint,
      // but active monitors that are not paused are assumed UP unless we have heartbeat data.
      // This will be improved once SSE provides status info.
      return !m.active;
    }).length;

    const canvas = document.createElement('canvas');
    canvas.width = 32;
    canvas.height = 32;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Draw base icon
    ctx.fillStyle = downCount > 0 ? '#dc3545' : '#5cdd8b';
    ctx.beginPath();
    ctx.arc(16, 16, 14, 0, Math.PI * 2);
    ctx.fill();

    if (downCount > 0) {
      ctx.fillStyle = '#fff';
      ctx.font = 'bold 16px sans-serif';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(downCount > 99 ? '99+' : String(downCount), 16, 17);
    } else {
      // Checkmark
      ctx.strokeStyle = '#fff';
      ctx.lineWidth = 3;
      ctx.lineCap = 'round';
      ctx.beginPath();
      ctx.moveTo(10, 16);
      ctx.lineTo(14, 21);
      ctx.lineTo(22, 11);
      ctx.stroke();
    }

    const link =
      (document.querySelector("link[rel*='icon']") as HTMLLinkElement) ||
      document.createElement('link');
    link.rel = 'icon';
    link.href = canvas.toDataURL();
    document.head.appendChild(link);
  }, [monitors]);

  return null;
}
