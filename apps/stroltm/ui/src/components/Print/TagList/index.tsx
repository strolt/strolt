import { Space } from "antd";

import { TagColored } from "components/TagColored";

export interface TagListProps {
  fallback?: React.ReactNode;
  value: string[];
}
export const TagList: React.FC<TagListProps> = ({ fallback, value }) => {
  if (value.length === 0 && fallback) {
    return <>{fallback}</>;
  }

  return (
    <Space wrap>
      {value.map((tag) => (
        <TagColored key={tag} value={tag} />
      ))}
    </Space>
  );
};
