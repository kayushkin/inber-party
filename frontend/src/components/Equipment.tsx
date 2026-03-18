import type { Equipment } from '../constants/equipment';
import { RARITY_COLORS } from '../constants/equipment';
import './Equipment.css';

interface EquipmentProps {
  equipment: Equipment[];
  className?: string;
}

export default function EquipmentComponent({ equipment, className = '' }: EquipmentProps) {

  if (equipment.length === 0) {
    return (
      <div className={`equipment-empty ${className}`}>
        <div className="equipment-slot empty">
          <span className="equipment-icon">🎒</span>
          <span className="equipment-empty-text">No gear equipped</span>
        </div>
      </div>
    );
  }

  // Group equipment by type
  const groupedEquipment = equipment.reduce((acc, item) => {
    if (!acc[item.type]) acc[item.type] = [];
    acc[item.type].push(item);
    return acc;
  }, {} as Record<string, Equipment[]>);

  const typeOrder = ['hat', 'weapon', 'armor', 'accessory', 'tool'] as const;
  const typeLabels: Record<string, string> = {
    hat: 'Headwear',
    weapon: 'Weapons',
    armor: 'Armor',
    accessory: 'Accessories',
    tool: 'Tools'
  };

  return (
    <div className={`equipment-grid ${className}`}>
      {typeOrder.map(type => {
        const items = groupedEquipment[type];
        if (!items || items.length === 0) return null;

        return (
          <div key={type} className="equipment-category">
            <div className="equipment-category-header">
              <h4 className="equipment-category-title">{typeLabels[type]}</h4>
              <div className="equipment-category-count">({items.length})</div>
            </div>
            <div className="equipment-items">
              {items.map(item => (
                <div
                  key={item.id}
                  className={`equipment-slot filled ${item.rarity}`}
                  style={{ borderColor: RARITY_COLORS[item.rarity] }}
                  title={`${item.name}: ${item.description}`}
                >
                  <div className="equipment-icon" style={{ color: RARITY_COLORS[item.rarity] }}>
                    {item.icon}
                  </div>
                  <div className="equipment-info">
                    <div className="equipment-name" style={{ color: RARITY_COLORS[item.rarity] }}>
                      {item.name}
                    </div>
                    <div className="equipment-type">{item.type}</div>
                    {(item.toolRequirement || item.skillRequirement) && (
                      <div className="equipment-requirements">
                        {item.toolRequirement && <span className="requirement">🔧 {item.toolRequirement}</span>}
                        {item.skillRequirement && <span className="requirement">⭐ {item.skillRequirement}</span>}
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        );
      })}
    </div>
  );
}