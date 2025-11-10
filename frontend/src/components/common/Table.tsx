import React from 'react';
import Button from './Button';

export interface Column {
  header: string;
  accessor: string;
  render?: (row: any) => React.ReactNode;
}

interface TableProps {
  columns: Column[];
  data: any[];
  totalCount: number;
  page: number;
  pageSize: number;
  onPageChange: (newPage: number) => void;
}

const Table: React.FC<TableProps> = ({ columns, data, totalCount, page, pageSize, onPageChange }) => {
  const totalPages = Math.ceil(totalCount / pageSize);

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full bg-white">
        <thead className="bg-gray-100">
          <tr>
            {columns.map((col) => (
              <th key={col.accessor} className="text-left py-3 px-4 uppercase font-semibold text-sm">
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="text-gray-700">
          {data.map((row, rowIndex) => (
            <tr key={rowIndex} className="border-b border-gray-200 hover:bg-gray-50">
              {columns.map((col) => (
                <td key={col.accessor} className="text-left py-3 px-4">
                  {col.render ? col.render(row) : row[col.accessor]}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      <div className="flex justify-between items-center mt-4">
        <span className="text-sm text-gray-600">
          Showing {Math.min((page - 1) * pageSize + 1, totalCount)} to {Math.min(page * pageSize, totalCount)} of {totalCount} results
        </span>
        <div className="flex items-center">
          <Button onClick={() => onPageChange(page - 1)} disabled={page <= 1} variant="secondary">
            Previous
          </Button>
          <span className="px-4 text-sm">
            Page {page} of {totalPages > 0 ? totalPages : 1}
          </span>
          <Button onClick={() => onPageChange(page + 1)} disabled={page >= totalPages} variant="secondary">
            Next
          </Button>
        </div>
      </div>
    </div>
  );
};

export default Table;

