const API_URL = "http://localhost:8080/upload";

import { Button } from "@/components/ui/button";
import { useTheme } from "@/hooks/useTheme";
import { cn } from "@/lib/utils";
import React, { useRef, useState, type FormEvent } from "react";
import { Input } from "./ui/input";

export default function SendFile() {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [recipient, setRecipient] = useState<string>("");
  const [error, setError] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [hasUploaded, setHasUploaded] = useState(false);
  const { setTheme, theme } = useTheme();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!selectedFile) {
      setError("Please select a file");
      return;
    }

    if (!recipient) {
      setError("Please enter a recipient");
      return;
    }

    const formData = new FormData();
    formData.append("file", selectedFile);
    formData.append("recipient", recipient);

    try {
      setIsUploading(true);
      const response = await fetch(API_URL, {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || "Upload failed");
      }

      const result = await response.json();
      setHasUploaded(true);
      console.log("Upload success:", result);
      // Reset after successful upload
      setSelectedFile(null);
      if (e.target instanceof HTMLFormElement) e.target.reset();
    } catch (err) {
      console.error("Upload error:", err);
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div className="h-svh w-full overflow-auto p-8">
      <div className="mx-auto flex max-w-lg items-center justify-between px-4 py-12">
        <span className="pr-10">🚀</span>
        <h1 className="text-2xl font-bold opacity-75">Send File</h1>
        <Button
          variant="ghost"
          className="text-foreground/50"
          onClick={() => {
            if (theme === "light") {
              setTheme("dark");
            } else {
              setTheme("light");
            }
          }}
        >
          <span className="dark:hidden">dark</span>
          <span className="hidden dark:inline">light</span>
        </Button>
      </div>

      <form onSubmit={handleSubmit} className="mx-auto w-full max-w-lg">
        <div className="flex gap-1">
          <div className="text-background bg-foreground/75 flex size-6 items-center justify-center rounded-full text-sm">
            1
          </div>
          <span>Choose File</span>
        </div>
        <div className="border-border m-3 border-l-2 p-4">
          <FileUploadButton
            selectedFile={selectedFile}
            setSelectedFile={setSelectedFile}
          />
        </div>
        <div className="flex gap-1">
          <div className="text-background bg-foreground/75 flex size-6 items-center justify-center rounded-full text-sm">
            2
          </div>
          <span>Send To</span>
        </div>
        <div className="border-border m-3 border-l-2 p-4">
          <Input
            value={recipient}
            onChange={(e) => setRecipient(e.target.value)}
            placeholder="Recipient ID"
          />
        </div>
        <div className="flex gap-1">
          <div className="text-background bg-foreground/75 flex size-6 items-center justify-center rounded-full text-sm">
            3
          </div>
          <span>Send</span>
        </div>
        <div className="m-3 flex flex-col items-center gap-4 border-l-2 border-transparent p-4">
          {
            <div
              className={cn("text-sm", {
                "text-red-500": error,
                "text-green-500": hasUploaded,
              })}
            >
              {error
                ? error
                : hasUploaded
                  ? "File sent successfully"
                  : "Send file to the daemon"}
            </div>
          }
          <Button className="cursor-pointer" type="submit">
            {isUploading ? "Sending to daemon..." : "Send to daemon"}
          </Button>
        </div>
      </form>
    </div>
  );
}

function FileUploadButton({
  selectedFile,
  setSelectedFile,
}: {
  selectedFile: File | null;
  setSelectedFile: (file: File | null) => void;
}) {
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleButtonClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      setSelectedFile(files[0]);
    } else {
      setSelectedFile(null);
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + " bytes";
    else if (bytes < 1048576) return (bytes / 1024).toFixed(1) + " KB";
    else return (bytes / 1048576).toFixed(1) + " MB";
  };

  return (
    <div className="flex max-w-md flex-col items-center gap-4 rounded-lg border p-6">
      <input
        type="file"
        ref={fileInputRef}
        onChange={handleFileChange}
        className="hidden"
        aria-label="File upload"
      />

      <Button
        onClick={handleButtonClick}
        variant="secondary"
        className="h-auto"
      >
        {selectedFile ? (
          <>
            <div className="">
              <p className="max-w-full truncate">{selectedFile.name}</p>
              <p>{formatFileSize(selectedFile.size)}</p>
            </div>
          </>
        ) : (
          "Upload File"
        )}
      </Button>
    </div>
  );
}
