import { Space } from "antd";

import { TagColored } from "components/TagColored";

export interface TagListProps {
  value: string[];
  fallback?: React.ReactNode;
}
export const TagList: React.FC<TagListProps> = ({ fallback, value }) => {
  if (!value.length && fallback) {
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
