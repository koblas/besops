import { useRef, useEffect, useState, useCallback } from 'react';
import { STATUS_COLORS, type StatusValue } from '../lib/constants';
import { formatDateTime } from '../lib/formatters';
import type { Heartbeat } from '../hooks/useHeartbeats';
import dayjs from 'dayjs';

interface HeartbeatBarProps {
  heartbeats: Heartbeat[];
  size?: 'small' | 'normal';
}

const BEAT_WIDTH = 10;
const BEAT_HEIGHT = 30;
const BEAT_PADDING = 4;
const HOVER_SCALE = 1.5;
const EMPTY_COLOR_LIGHT = '#f0f8ff';
const EMPTY_COLOR_DARK = '#848484';

export function HeartbeatBar({ heartbeats, size = 'normal' }: HeartbeatBarProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [maxBeats, setMaxBeats] = useState(0);
  const [hoverIndex, setHoverIndex] = useState<number | null>(null);
  const [tooltipPos, setTooltipPos] = useState<{ x: number; y: number } | null>(null);

  const beatWidth = size === 'small' ? 5 : BEAT_WIDTH;
  const beatHeight = size === 'small' ? 16 : BEAT_HEIGHT;
  const beatPadding = size === 'small' ? 2 : BEAT_PADDING;
  const beatFullWidth = beatWidth + beatPadding * 2;
  const canvasHeight = beatHeight * HOVER_SCALE;

  useEffect(() => {
    if (!containerRef.current) return;
    const observer = new ResizeObserver(entries => {
      const width = entries[0].contentRect.width;
      setMaxBeats(Math.floor(width / beatFullWidth));
    });
    observer.observe(containerRef.current);
    return () => observer.disconnect();
  }, [beatFullWidth]);

  const visibleBeats = heartbeats.slice(-maxBeats);
  const emptySlots = maxBeats - visibleBeats.length;

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || maxBeats === 0) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const dpr = window.devicePixelRatio || 1;
    const width = maxBeats * beatFullWidth;
    canvas.width = width * dpr;
    canvas.height = canvasHeight * dpr;
    canvas.style.width = `${width}px`;
    canvas.style.height = `${canvasHeight}px`;
    ctx.scale(dpr, dpr);
    ctx.clearRect(0, 0, width, canvasHeight);

    const isDark = document.documentElement.classList.contains('dark') ||
      document.body.getAttribute('data-theme') === 'dark';
    const emptyColor = isDark ? EMPTY_COLOR_DARK : EMPTY_COLOR_LIGHT;
    const centerY = canvasHeight / 2;

    for (let i = 0; i < maxBeats; i++) {
      const beatIdx = i - emptySlots;
      const isHovered = beatIdx >= 0 && beatIdx === hoverIndex;

      let w = beatWidth;
      let h = beatHeight;
      let x = i * beatFullWidth + beatPadding;
      let y = centerY - h / 2;

      if (isHovered && beatIdx >= 0) {
        w *= HOVER_SCALE;
        h *= HOVER_SCALE;
        x = i * beatFullWidth + beatPadding - (w - beatWidth) / 2;
        y = centerY - h / 2;
      }

      let color: string;
      if (beatIdx < 0) {
        color = emptyColor;
      } else {
        const beat = visibleBeats[beatIdx];
        color = STATUS_COLORS[beat.status as StatusValue] || emptyColor;
      }

      const radius = w / 2;
      ctx.fillStyle = color;
      ctx.beginPath();
      ctx.roundRect(x, y, w, h, radius);
      ctx.fill();
    }
  }, [heartbeats, maxBeats, hoverIndex, canvasHeight, beatWidth, beatHeight, beatPadding, beatFullWidth, emptySlots, visibleBeats]);

  const handleMouseMove = useCallback(
    (e: React.MouseEvent) => {
      const rect = canvasRef.current?.getBoundingClientRect();
      if (!rect) return;
      const x = e.clientX - rect.left;
      const idx = Math.floor(x / beatFullWidth);
      const beatIdx = idx - emptySlots;
      if (beatIdx >= 0 && beatIdx < visibleBeats.length) {
        setHoverIndex(beatIdx);
        setTooltipPos({ x: e.clientX, y: e.clientY });
      } else {
        setHoverIndex(null);
        setTooltipPos(null);
      }
    },
    [visibleBeats.length, emptySlots, beatFullWidth],
  );

  const hoveredBeat = hoverIndex !== null ? visibleBeats[hoverIndex] : null;

  const timeSinceFirst = (() => {
    if (size === 'small' || visibleBeats.length < 5) return null;
    const first = visibleBeats[0];
    if (!first) return null;
    const minutes = dayjs().diff(dayjs(first.time), 'minute');
    if (minutes > 60) return `${Math.floor(minutes / 60)}h`;
    return `${minutes}m`;
  })();

  const timeSinceLast = (() => {
    if (size === 'small' || visibleBeats.length < 5) return null;
    const last = visibleBeats[visibleBeats.length - 1];
    if (!last) return null;
    const seconds = dayjs().diff(dayjs(last.time), 'second');
    if (seconds < 120) return 'Now';
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    return `${Math.floor(seconds / 3600)}h ago`;
  })();

  const wrapPadding = (beatHeight * HOVER_SCALE - beatHeight) / 2;

  return (
    <div ref={containerRef} style={{ width: '100%', overflow: 'hidden' }}>
      <div style={{ padding: `${wrapPadding}px ${(beatWidth * HOVER_SCALE - beatWidth) / 2}px` }}>
        <canvas
          ref={canvasRef}
          role="img"
          aria-label={`Heartbeat history: ${visibleBeats.length} checks shown`}
          style={{ display: 'block' }}
          onMouseMove={handleMouseMove}
          onMouseLeave={() => { setHoverIndex(null); setTooltipPos(null); }}
        />
      </div>
      {hoveredBeat && tooltipPos && (
        <div
          style={{
            position: 'fixed',
            left: tooltipPos.x,
            top: tooltipPos.y - 8,
            transform: 'translate(-50%, -100%)',
            background: 'rgba(0, 0, 0, 0.85)',
            color: '#fff',
            padding: '6px 10px',
            borderRadius: 6,
            fontSize: 12,
            whiteSpace: 'nowrap',
            pointerEvents: 'none',
            zIndex: 1000,
          }}
        >
          <div>{formatDateTime(hoveredBeat.time)}</div>
          <div>
            {hoveredBeat.msg || (hoveredBeat.status === 1 ? 'Up' : 'Down')}
            {hoveredBeat.ping != null && ` (${hoveredBeat.ping}ms)`}
          </div>
        </div>
      )}
      {timeSinceFirst && timeSinceLast && (
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            fontSize: 12,
            color: '#888',
            marginTop: 2,
            marginLeft: emptySlots * beatFullWidth,
          }}
        >
          <span>{timeSinceFirst}</span>
          <span>{timeSinceLast}</span>
        </div>
      )}
    </div>
  );
}
