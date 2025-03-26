import React from 'react';
import { Link } from 'react-router-dom';
import { ExternalLink, Edit2, Trash2 } from 'react-feather';
import RouteFeatureBadges from './RouteFeatureBadges';

export default function RouteTableRow({ route, onDelete }) {
  return (
    <tr>
      <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">
        {route.Name}
      </td>
      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
        {route.ListenPath}
      </td>
      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
        <span className="truncate max-w-xs block">{route.UpstreamURL}</span>
      </td>
      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
        {route.Methods?.includes('*') ? 'ALL' : route.Methods?.join(', ')}
      </td>
      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
        <RouteFeatureBadges route={route} />
      </td>
      <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
        <div className="flex justify-end space-x-2">
          <Link
            to={`/routes/${route.Name}`}
            className="text-indigo-600 hover:text-indigo-900 inline-flex items-center"
          >
            <ExternalLink className="w-4 h-4 mr-1" />
            View
          </Link>
          <Link
            to={`/routes/${route.Name}/edit`}
            className="text-indigo-600 hover:text-indigo-900 inline-flex items-center"
          >
            <Edit2 className="w-4 h-4 mr-1" />
            Edit
          </Link>
          <button
            onClick={() => onDelete(route)}
            className="text-red-600 hover:text-red-900 inline-flex items-center"
          >
            <Trash2 className="w-4 h-4 mr-1" />
            Delete
          </button>
        </div>
      </td>
    </tr>
  );
}