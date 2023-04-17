import { Space } from "antd";

import { TagColored } from "components/TagColored";

export interface TagListProps {
  value: string[];
}
export const TagList: React.FC<TagListProps> = ({ value }) => {
  return (
    <Space wrap>
      {value.map((tag) => (
        <TagColored key={tag} value={tag} />
      ))}
    </Space>
  );
};
