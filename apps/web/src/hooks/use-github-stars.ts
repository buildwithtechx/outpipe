import { useQuery } from '@tanstack/react-query';
import { getGitHubStarCount } from '#/lib/github';

export function useGitHubStars() {
  return useQuery({
    queryKey: ['github', 'buildwithtechx', 'outpipe', 'stars'],
    queryFn: getGitHubStarCount,
    staleTime: 1000 * 60 * 5,
    refetchInterval: 1000 * 60 * 5,
    retry: 1,
  });
}
