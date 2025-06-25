import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { Check, Copy } from 'lucide-react';
import { useToast } from '@/components/ui/use-toast';

interface CodeBlockProps {
  code: string;
  className?: string;
}

export const CodeBlock: React.FC<CodeBlockProps> = ({ code, className = "" }) => {
  const [copied, setCopied] = useState(false);
  const { toast } = useToast();

    const copyToClipboard = async () => {
    try {
      // Try modern clipboard API first
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(code);
      } else {
        // Fallback for older browsers or non-secure contexts
        const textArea = document.createElement('textarea');
        textArea.value = code;
        textArea.style.position = 'fixed';
        textArea.style.opacity = '0';
        textArea.style.left = '-999999px';
        textArea.style.top = '-999999px';
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
      }

      setCopied(true);
      toast({
        description: "Command copied to clipboard",
        duration: 2000,
      });

      // Reset the copied state after 2 seconds
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Copy failed:', err);
      toast({
        description: "Failed to copy to clipboard. Please copy manually.",
        variant: "destructive",
        duration: 3000,
      });
    }
  };

  return (
    <div className={`relative bg-gray-900 rounded-lg p-4 group ${className}`}>
      <code className="text-green-400 text-sm block pr-12">{code}</code>

      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              className="absolute top-2 right-2 h-8 w-8 p-0 opacity-0 group-hover:opacity-100 transition-opacity hover:bg-gray-800"
              onClick={copyToClipboard}
            >
              {copied ? (
                <Check className="h-4 w-4 text-green-400" />
              ) : (
                <Copy className="h-4 w-4 text-gray-400 hover:text-white" />
              )}
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>{copied ? 'Copied!' : 'Copy command'}</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
};
