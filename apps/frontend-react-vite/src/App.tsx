import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import HomePage from './routes/index'
import { Toaster } from '@/components/ui/toaster'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <HomePage />
      <Toaster />
    </QueryClientProvider>
  )
}
