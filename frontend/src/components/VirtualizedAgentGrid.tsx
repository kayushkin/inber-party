import React, { memo, useMemo, useState, useEffect, useRef } from 'react';
import AgentCard from './AgentCard';
import type { RPGAgent } from '../store';
import './VirtualizedAgentGrid.css';

interface VirtualizedAgentGridProps {
  agents: RPGAgent[];
  selectedAgent?: string | null;
  onAgentClick: (agent: RPGAgent) => void;
  itemMinWidth?: number;
  gap?: number;
  className?: string;
}

interface VisibleRange {
  start: number;
  end: number;
}

const VirtualizedAgentGrid = memo<VirtualizedAgentGridProps>(({
  agents,
  selectedAgent,
  onAgentClick,
  itemMinWidth = 280,
  gap = 16,
  className = ''
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [containerDimensions, setContainerDimensions] = useState({ width: 0, height: 0 });
  const [scrollTop, setScrollTop] = useState(0);
  const [visibleRange, setVisibleRange] = useState<VisibleRange>({ start: 0, end: 10 });
  
  // Calculate grid layout
  const { columnsPerRow, itemWidth, totalRows, itemHeight } = useMemo(() => {
    if (containerDimensions.width === 0) {
      return { columnsPerRow: 1, itemWidth: itemMinWidth, totalRows: agents.length, itemHeight: 200 };
    }
    
    const availableWidth = containerDimensions.width - gap;
    const cols = Math.max(1, Math.floor(availableWidth / (itemMinWidth + gap)));
    const actualItemWidth = (availableWidth - (cols - 1) * gap) / cols;
    const rows = Math.ceil(agents.length / cols);
    
    return {
      columnsPerRow: cols,
      itemWidth: actualItemWidth,
      totalRows: rows,
      itemHeight: 180 // Estimated height of agent card
    };
  }, [agents.length, containerDimensions.width, itemMinWidth, gap]);

  // Handle container resize
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setContainerDimensions({
          width: entry.contentRect.width,
          height: entry.contentRect.height
        });
      }
    });

    resizeObserver.observe(container);
    return () => resizeObserver.disconnect();
  }, []);

  // Calculate visible range based on scroll position
  useEffect(() => {
    const rowHeight = itemHeight + gap;
    const visibleRows = Math.ceil(containerDimensions.height / rowHeight);
    const bufferRows = 2; // Render extra rows for smooth scrolling
    
    const startRow = Math.max(0, Math.floor(scrollTop / rowHeight) - bufferRows);
    const endRow = Math.min(totalRows, startRow + visibleRows + bufferRows * 2);
    
    const startIndex = startRow * columnsPerRow;
    const endIndex = Math.min(agents.length, endRow * columnsPerRow);
    
    setVisibleRange({ start: startIndex, end: endIndex });
  }, [scrollTop, containerDimensions.height, totalRows, columnsPerRow, itemHeight, gap, agents.length]);

  // Handle scroll
  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    setScrollTop(e.currentTarget.scrollTop);
  };

  // Create visible items
  const visibleItems = useMemo(() => {
    const items: React.ReactNode[] = [];
    
    for (let i = visibleRange.start; i < visibleRange.end && i < agents.length; i++) {
      const agent = agents[i];
      const row = Math.floor(i / columnsPerRow);
      const col = i % columnsPerRow;
      
      const x = col * (itemWidth + gap);
      const y = row * (itemHeight + gap);
      
      items.push(
        <div
          key={agent.id}
          className="virtual-grid-item"
          style={{
            position: 'absolute',
            left: x,
            top: y,
            width: itemWidth,
            height: itemHeight
          }}
        >
          <AgentCard
            agent={agent}
            isSelected={selectedAgent === agent.id}
            onClick={onAgentClick}
            compact
          />
        </div>
      );
    }
    
    return items;
  }, [visibleRange, agents, columnsPerRow, itemWidth, itemHeight, gap, selectedAgent, onAgentClick]);

  // Total height for scrollbar
  const totalHeight = totalRows * (itemHeight + gap) - gap;

  // Fallback to non-virtualized grid for small lists
  if (agents.length <= 12) {
    return (
      <div className={`agents-grid performance ${className}`}>
        {agents.map(agent => (
          <AgentCard
            key={agent.id}
            agent={agent}
            isSelected={selectedAgent === agent.id}
            onClick={onAgentClick}
            compact
          />
        ))}
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className={`virtual-grid-container ${className}`}
      onScroll={handleScroll}
    >
      <div
        className="virtual-grid-content"
        style={{ height: totalHeight, position: 'relative' }}
      >
        {visibleItems}
      </div>
    </div>
  );
});

VirtualizedAgentGrid.displayName = 'VirtualizedAgentGrid';

export default VirtualizedAgentGrid;