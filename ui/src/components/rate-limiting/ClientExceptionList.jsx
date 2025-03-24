import React from 'react';
import { Trash2 } from 'react-feather';

export default function ClientExceptionList({ exceptions = [], onRemove }) {
  if (exceptions.length === 0) {
    return null;
  }

  return (
    <div className="mt-3">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Client Key
            </th>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Limit
            </th>
            <th scope="col" className="relative px-6 py-3">
              <span className="sr-only">Actions</span>
            </th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {exceptions.map((exception, index) => (
            <tr key={index}>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {exception.key}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {exception.limit}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button
                  type="button"
                  onClick={() => onRemove(index)}
                  className="text-red-600 hover:text-red-900"
                  aria-label={`Remove exception for ${exception.key}`}
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}