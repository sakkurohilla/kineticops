// Date utility functions for proper timezone handling

export const formatTimestamp = (timestamp: string | Date): string => {
  try {
    const date = new Date(timestamp);
    
    // Check if date is valid
    if (isNaN(date.getTime())) {
      return 'Invalid Date';
    }
    
    // Format in local timezone
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: true
    });
  } catch (error) {
    console.error('Error formatting timestamp:', error);
    return 'Invalid Date';
  }
};

export const formatRelativeTime = (timestamp: string | Date): string => {
  try {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMinutes = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    
    if (diffMinutes < 1) return 'Just now';
    if (diffMinutes < 60) return `${diffMinutes}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    
    return formatTimestamp(timestamp);
  } catch (error) {
    return 'Unknown';
  }
};

export const isOnline = (lastSeen: string | Date, thresholdMinutes: number = 5): boolean => {
  try {
    const date = new Date(lastSeen);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMinutes = diffMs / (1000 * 60);
    
    return diffMinutes <= thresholdMinutes;
  } catch (error) {
    return false;
  }
};