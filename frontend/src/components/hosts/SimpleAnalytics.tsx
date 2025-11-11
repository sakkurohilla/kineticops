import React from 'react'
import Card from '../common/Card'

interface Props {
  cpu: number
  memoryUsedBytes: number
  memoryTotalBytes: number
  diskUsedBytes: number
  diskTotalBytes: number
}

const toGB = (bytes: number) => (bytes / (1024 * 1024 * 1024))

const SimpleAnalytics: React.FC<Props> = ({ cpu, memoryUsedBytes, memoryTotalBytes, diskUsedBytes, diskTotalBytes }) => {
  const memGB = toGB(memoryUsedBytes)
  const memTotalGB = toGB(memoryTotalBytes)
  const diskGB = toGB(diskUsedBytes)
  const diskTotalGB = toGB(diskTotalBytes)

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
      <Card>
        <p className="text-xs text-gray-500">CPU</p>
        <div className="text-2xl font-bold mt-2">{cpu.toFixed(1)}%</div>
      </Card>

      <Card>
        <p className="text-xs text-gray-500">Memory</p>
        <div className="text-2xl font-bold mt-2">{memGB.toFixed(2)} GB</div>
        <p className="text-sm text-gray-500">{memGB.toFixed(2)} / {memTotalGB.toFixed(2)} GB</p>
      </Card>

      <Card>
        <p className="text-xs text-gray-500">Disk</p>
        <div className="text-2xl font-bold mt-2">{diskGB.toFixed(2)} GB</div>
        <p className="text-sm text-gray-500">{diskGB.toFixed(2)} / {diskTotalGB.toFixed(2)} GB</p>
      </Card>
    </div>
  )
}

export default SimpleAnalytics
