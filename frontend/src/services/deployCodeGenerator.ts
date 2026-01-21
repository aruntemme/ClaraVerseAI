// Code generator for deployment API calls across multiple languages
import type { VariableConfig } from '@/types/agent';

export type Language =
  | 'curl'
  | 'python'
  | 'javascript-axios'
  | 'javascript-fetch'
  | 'javascript-ajax'
  | 'go';

export type InputType = 'text' | 'file' | 'json';

export interface InputField {
  name: string;
  type: InputType;
  value?: string;
  required: boolean;
}

export interface CodeGenOptions {
  language: Language;
  triggerUrl: string;
  statusUrl: string;
  apiKey: string;
  agentId: string;
  startBlockConfig: VariableConfig | null;
}

/**
 * Detect input configuration from Start block
 */
export function detectInputFromStartBlock(config: VariableConfig | null): {
  inputType: InputType;
  hasInput: boolean;
  sampleValue: string;
} {
  if (!config) {
    return { inputType: 'text', hasInput: false, sampleValue: '"Hello, World!"' };
  }

  const inputType = config.inputType || 'text';
  const hasInput = !!(config.defaultValue || config.fileValue || config.jsonValue);

  let sampleValue = '"Hello, World!"';

  if (inputType === 'file') {
    sampleValue = 'FILE_ID_HERE'; // Placeholder for file upload
  } else if (inputType === 'json') {
    sampleValue = config.jsonValue ? JSON.stringify(config.jsonValue, null, 2) : '{"key": "value"}';
  } else if (config.defaultValue) {
    sampleValue = `"${config.defaultValue}"`;
  }

  return { inputType, hasInput, sampleValue };
}

/**
 * Generate code for uploading a file (for file input type)
 */
function generateFileUploadCode(language: Language, apiKey: string, baseUrl: string): string {
  const uploadUrl = `${baseUrl}/api/external/upload`;

  switch (language) {
    case 'curl':
      return `# 1. Upload file first
FILE_RESPONSE=$(curl -X POST \\
  "${uploadUrl}" \\
  -H "X-API-Key: ${apiKey}" \\
  -F "file=@/path/to/your/file.pdf")

FILE_ID=$(echo $FILE_RESPONSE | jq -r '.file_id')
FILENAME=$(echo $FILE_RESPONSE | jq -r '.filename')
MIME_TYPE=$(echo $FILE_RESPONSE | jq -r '.mime_type')`;

    case 'python':
      return `# 1. Upload file first
import requests

upload_url = "${uploadUrl}"
headers = {"X-API-Key": "${apiKey}"}

with open('/path/to/your/file.pdf', 'rb') as f:
    files = {'file': f}
    upload_response = requests.post(upload_url, headers=headers, files=files)
    file_data = upload_response.json()

file_id = file_data['file_id']
filename = file_data['filename']
mime_type = file_data['mime_type']`;

    case 'javascript-axios':
      return `// 1. Upload file first
const FormData = require('form-data');
const fs = require('fs');
const axios = require('axios');

const uploadUrl = '${uploadUrl}';
const formData = new FormData();
formData.append('file', fs.createReadStream('/path/to/your/file.pdf'));

const uploadResponse = await axios.post(uploadUrl, formData, {
  headers: {
    'X-API-Key': '${apiKey}',
    ...formData.getHeaders()
  }
});

const fileData = uploadResponse.data;`;

    case 'javascript-fetch':
      return `// 1. Upload file first
const uploadUrl = '${uploadUrl}';
const formData = new FormData();
formData.append('file', fileInput.files[0]); // From <input type="file" />

const uploadResponse = await fetch(uploadUrl, {
  method: 'POST',
  headers: { 'X-API-Key': '${apiKey}' },
  body: formData
});

const fileData = await uploadResponse.json();`;

    case 'javascript-ajax':
      return `// 1. Upload file first (jQuery)
var formData = new FormData();
formData.append('file', $('#fileInput')[0].files[0]);

var fileData;
$.ajax({
  url: '${uploadUrl}',
  type: 'POST',
  data: formData,
  headers: { 'X-API-Key': '${apiKey}' },
  processData: false,
  contentType: false,
  async: false,
  success: function(response) {
    fileData = response;
  }
});`;

    case 'go':
      return `// 1. Upload file first
package main

import (
    "bytes"
    "encoding/json"
    "io"
    "mime/multipart"
    "net/http"
    "os"
)

file, _ := os.Open("/path/to/your/file.pdf")
defer file.Close()

body := &bytes.Buffer{}
writer := multipart.NewWriter(body)
part, _ := writer.CreateFormFile("file", "file.pdf")
io.Copy(part, file)
writer.Close()

req, _ := http.NewRequest("POST", "${uploadUrl}", body)
req.Header.Set("X-API-Key", "${apiKey}")
req.Header.Set("Content-Type", writer.FormDataContentType())

client := &http.Client{}
uploadResp, _ := client.Do(req)
defer uploadResp.Body.Close()

var fileData map[string]interface{}
json.NewDecoder(uploadResp.Body).Decode(&fileData)`;

    default:
      return '';
  }
}

/**
 * Generate trigger request body based on input type
 */
function generateInputBody(
  inputType: InputType,
  sampleValue: string,
  isFileUpload: boolean
): string {
  if (inputType === 'file' && isFileUpload) {
    return `{
  "input": {
    "file_id": FILE_ID,
    "filename": FILENAME,
    "mime_type": MIME_TYPE
  }
}`;
  } else if (inputType === 'json') {
    return `{
  "input": ${sampleValue}
}`;
  } else {
    return `{
  "input": ${sampleValue}
}`;
  }
}

/**
 * Generate complete code with polling
 */
export function generateDeploymentCode(options: CodeGenOptions): string {
  const { language, triggerUrl, statusUrl, apiKey, startBlockConfig } = options;
  const { inputType, sampleValue } = detectInputFromStartBlock(startBlockConfig);

  const baseUrl = triggerUrl.split('/api/trigger')[0];
  const hasFileUpload = inputType === 'file';

  switch (language) {
    case 'curl':
      return generateCurlCode(
        triggerUrl,
        statusUrl,
        apiKey,
        baseUrl,
        inputType,
        sampleValue,
        hasFileUpload
      );
    case 'python':
      return generatePythonCode(
        triggerUrl,
        statusUrl,
        apiKey,
        baseUrl,
        inputType,
        sampleValue,
        hasFileUpload
      );
    case 'javascript-axios':
      return generateJSAxiosCode(
        triggerUrl,
        statusUrl,
        apiKey,
        baseUrl,
        inputType,
        sampleValue,
        hasFileUpload
      );
    case 'javascript-fetch':
      return generateJSFetchCode(
        triggerUrl,
        statusUrl,
        apiKey,
        baseUrl,
        inputType,
        sampleValue,
        hasFileUpload
      );
    case 'javascript-ajax':
      return generateJSAjaxCode(
        triggerUrl,
        statusUrl,
        apiKey,
        baseUrl,
        inputType,
        sampleValue,
        hasFileUpload
      );
    case 'go':
      return generateGoCode(
        triggerUrl,
        statusUrl,
        apiKey,
        baseUrl,
        inputType,
        sampleValue,
        hasFileUpload
      );
    default:
      return '';
  }
}

function generateCurlCode(
  triggerUrl: string,
  statusUrl: string,
  apiKey: string,
  baseUrl: string,
  inputType: InputType,
  sampleValue: string,
  hasFileUpload: boolean
): string {
  const fileUploadSection = hasFileUpload
    ? generateFileUploadCode('curl', apiKey, baseUrl) + '\n\n'
    : '';
  const inputBody = generateInputBody(inputType, sampleValue, hasFileUpload);

  return `${fileUploadSection}# 2. Trigger workflow execution
EXECUTION_ID=$(curl -X POST \\
  "${triggerUrl}" \\
  -H "X-API-Key: ${apiKey}" \\
  -H "Content-Type: application/json" \\
  -d '${inputBody}' \\
  | jq -r '.executionId')

echo "Execution started: $EXECUTION_ID"

# 3. Poll for completion
while true; do
  RESPONSE=$(curl -s -X GET \\
    "${statusUrl.replace(':executionId', '$EXECUTION_ID')}" \\
    -H "X-API-Key: ${apiKey}")

  STATUS=$(echo $RESPONSE | jq -r '.status')

  if [ "$STATUS" = "completed" ]; then
    echo "✅ Execution completed!"
    echo "Output:"
    echo $RESPONSE | jq '.output'
    break
  elif [ "$STATUS" = "failed" ]; then
    echo "❌ Execution failed!"
    echo $RESPONSE | jq '.error'
    break
  elif [ "$STATUS" = "partial" ]; then
    echo "⚠️  Execution completed with partial results"
    echo "Output:"
    echo $RESPONSE | jq '.output'
    echo "Error:"
    echo $RESPONSE | jq '.error'
    break
  fi

  echo "⏳ Status: $STATUS - polling again in 2s..."
  sleep 2
done`;
}

function generatePythonCode(
  triggerUrl: string,
  statusUrl: string,
  apiKey: string,
  baseUrl: string,
  inputType: InputType,
  sampleValue: string,
  hasFileUpload: boolean
): string {
  const fileUploadSection = hasFileUpload
    ? generateFileUploadCode('python', apiKey, baseUrl) + '\n\n'
    : '';
  const inputBody = hasFileUpload
    ? `{
    "input": {
        "file_id": file_id,
        "filename": filename,
        "mime_type": mime_type
    }
}`
    : inputType === 'json'
      ? `{"input": ${sampleValue}}`
      : `{"input": ${sampleValue}}`;

  return `${fileUploadSection}# 2. Trigger workflow execution
import requests
import time
import json

trigger_url = "${triggerUrl}"
headers = {
    "X-API-Key": "${apiKey}",
    "Content-Type": "application/json"
}

# Trigger execution
payload = ${inputBody}
response = requests.post(trigger_url, headers=headers, json=payload)
data = response.json()

execution_id = data['executionId']
print(f"Execution started: {execution_id}")

# 3. Poll for completion
status_url = "${statusUrl.replace(':executionId', '{execution_id}')}"
while True:
    response = requests.get(status_url, headers=headers)
    execution = response.json()

    status = execution['status']

    if status == 'completed':
        print("✅ Execution completed!")
        print("Output:", json.dumps(execution.get('output'), indent=2))
        break
    elif status == 'failed':
        print("❌ Execution failed!")
        print("Error:", execution.get('error'))
        break
    elif status == 'partial':
        print("⚠️  Execution completed with partial results")
        print("Output:", json.dumps(execution.get('output'), indent=2))
        print("Error:", execution.get('error'))
        break

    print(f"⏳ Status: {status} - polling again in 2s...")
    time.sleep(2)`;
}

function generateJSAxiosCode(
  triggerUrl: string,
  statusUrl: string,
  apiKey: string,
  baseUrl: string,
  inputType: InputType,
  sampleValue: string,
  hasFileUpload: boolean
): string {
  const fileUploadSection = hasFileUpload
    ? generateFileUploadCode('javascript-axios', apiKey, baseUrl) + '\n\n'
    : '';
  const inputBody = hasFileUpload
    ? `{
  input: {
    file_id: fileData.file_id,
    filename: fileData.filename,
    mime_type: fileData.mime_type
  }
}`
    : inputType === 'json'
      ? `{ input: ${sampleValue} }`
      : `{ input: ${sampleValue} }`;

  return `${fileUploadSection}// 2. Trigger workflow execution
const axios = require('axios');

const triggerUrl = '${triggerUrl}';
const headers = {
  'X-API-Key': '${apiKey}',
  'Content-Type': 'application/json'
};

// Trigger execution
const triggerResponse = await axios.post(triggerUrl, ${inputBody}, { headers });
const { executionId } = triggerResponse.data;

console.log(\`Execution started: \${executionId}\`);

// 3. Poll for completion
const statusUrl = '${statusUrl.replace(':executionId', '${executionId}')}';
while (true) {
  const statusResponse = await axios.get(statusUrl, { headers });
  const execution = statusResponse.data;

  if (execution.status === 'completed') {
    console.log('✅ Execution completed!');
    console.log('Output:', JSON.stringify(execution.output, null, 2));
    break;
  } else if (execution.status === 'failed') {
    console.log('❌ Execution failed!');
    console.log('Error:', execution.error);
    break;
  } else if (execution.status === 'partial') {
    console.log('⚠️  Execution completed with partial results');
    console.log('Output:', JSON.stringify(execution.output, null, 2));
    console.log('Error:', execution.error);
    break;
  }

  console.log(\`⏳ Status: \${execution.status} - polling again in 2s...\`);
  await new Promise(resolve => setTimeout(resolve, 2000));
}`;
}

function generateJSFetchCode(
  triggerUrl: string,
  statusUrl: string,
  apiKey: string,
  baseUrl: string,
  inputType: InputType,
  sampleValue: string,
  hasFileUpload: boolean
): string {
  const fileUploadSection = hasFileUpload
    ? generateFileUploadCode('javascript-fetch', apiKey, baseUrl) + '\n\n'
    : '';
  const inputBody = hasFileUpload
    ? `{
  input: {
    file_id: fileData.file_id,
    filename: fileData.filename,
    mime_type: fileData.mime_type
  }
}`
    : inputType === 'json'
      ? `{ input: ${sampleValue} }`
      : `{ input: ${sampleValue} }`;

  return `${fileUploadSection}// 2. Trigger workflow execution
const triggerUrl = '${triggerUrl}';
const headers = {
  'X-API-Key': '${apiKey}',
  'Content-Type': 'application/json'
};

// Trigger execution
const triggerResponse = await fetch(triggerUrl, {
  method: 'POST',
  headers: headers,
  body: JSON.stringify(${inputBody})
});

const { executionId } = await triggerResponse.json();
console.log(\`Execution started: \${executionId}\`);

// 3. Poll for completion
const statusUrl = '${statusUrl.replace(':executionId', '${executionId}')}';
while (true) {
  const statusResponse = await fetch(statusUrl, { headers });
  const execution = await statusResponse.json();

  if (execution.status === 'completed') {
    console.log('✅ Execution completed!');
    console.log('Output:', JSON.stringify(execution.output, null, 2));
    break;
  } else if (execution.status === 'failed') {
    console.log('❌ Execution failed!');
    console.log('Error:', execution.error);
    break;
  } else if (execution.status === 'partial') {
    console.log('⚠️  Execution completed with partial results');
    console.log('Output:', JSON.stringify(execution.output, null, 2));
    console.log('Error:', execution.error);
    break;
  }

  console.log(\`⏳ Status: \${execution.status} - polling again in 2s...\`);
  await new Promise(resolve => setTimeout(resolve, 2000));
}`;
}

function generateJSAjaxCode(
  triggerUrl: string,
  statusUrl: string,
  apiKey: string,
  baseUrl: string,
  inputType: InputType,
  sampleValue: string,
  hasFileUpload: boolean
): string {
  const fileUploadSection = hasFileUpload
    ? generateFileUploadCode('javascript-ajax', apiKey, baseUrl) + '\n\n'
    : '';
  const inputBody = hasFileUpload
    ? `{
  input: {
    file_id: fileData.file_id,
    filename: fileData.filename,
    mime_type: fileData.mime_type
  }
}`
    : inputType === 'json'
      ? `{ input: ${sampleValue} }`
      : `{ input: ${sampleValue} }`;

  return `${fileUploadSection}// 2. Trigger workflow execution (jQuery)
var triggerUrl = '${triggerUrl}';
var headers = { 'X-API-Key': '${apiKey}' };

// Trigger execution
$.ajax({
  url: triggerUrl,
  type: 'POST',
  headers: headers,
  contentType: 'application/json',
  data: JSON.stringify(${inputBody}),
  success: function(response) {
    var executionId = response.executionId;
    console.log('Execution started: ' + executionId);

    // 3. Poll for completion
    var statusUrl = '${statusUrl.replace(':executionId', '' + 'executionId')}';
    var pollInterval = setInterval(function() {
      $.ajax({
        url: statusUrl,
        type: 'GET',
        headers: headers,
        success: function(execution) {
          if (execution.status === 'completed') {
            console.log('✅ Execution completed!');
            console.log('Output:', JSON.stringify(execution.output, null, 2));
            clearInterval(pollInterval);
          } else if (execution.status === 'failed') {
            console.log('❌ Execution failed!');
            console.log('Error:', execution.error);
            clearInterval(pollInterval);
          } else if (execution.status === 'partial') {
            console.log('⚠️  Execution completed with partial results');
            console.log('Output:', JSON.stringify(execution.output, null, 2));
            console.log('Error:', execution.error);
            clearInterval(pollInterval);
          } else {
            console.log('⏳ Status: ' + execution.status + ' - polling...');
          }
        }
      });
    }, 2000);
  }
});`;
}

function generateGoCode(
  triggerUrl: string,
  statusUrl: string,
  apiKey: string,
  baseUrl: string,
  inputType: InputType,
  sampleValue: string,
  hasFileUpload: boolean
): string {
  const fileUploadSection = hasFileUpload
    ? generateFileUploadCode('go', apiKey, baseUrl) + '\n\n'
    : '';
  const inputBody = hasFileUpload
    ? `map[string]interface{}{
		"input": map[string]interface{}{
			"file_id":   fileData["file_id"],
			"filename":  fileData["filename"],
			"mime_type": fileData["mime_type"],
		},
	}`
    : inputType === 'json'
      ? `map[string]interface{}{"input": ${sampleValue}}`
      : `map[string]interface{}{"input": ${sampleValue}}`;

  return `${fileUploadSection}// 2. Trigger workflow execution
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func main() {
	triggerURL := "${triggerUrl}"
	apiKey := "${apiKey}"

	// Trigger execution
	payload := ${inputBody}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", triggerURL, bytes.NewBuffer(jsonData))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var triggerResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&triggerResp)

	executionID := triggerResp["executionId"].(string)
	fmt.Printf("Execution started: %s\\n", executionID)

	// 3. Poll for completion
	statusURL := strings.Replace("${statusUrl}", ":executionId", executionID, 1)

	for {
		req, _ := http.NewRequest("GET", statusURL, nil)
		req.Header.Set("X-API-Key", apiKey)

		resp, _ := client.Do(req)
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		var execution map[string]interface{}
		json.Unmarshal(body, &execution)

		status := execution["status"].(string)

		if status == "completed" {
			fmt.Println("✅ Execution completed!")
			output, _ := json.MarshalIndent(execution["output"], "", "  ")
			fmt.Println("Output:", string(output))
			break
		} else if status == "failed" {
			fmt.Println("❌ Execution failed!")
			fmt.Println("Error:", execution["error"])
			break
		} else if status == "partial" {
			fmt.Println("⚠️  Execution completed with partial results")
			output, _ := json.MarshalIndent(execution["output"], "", "  ")
			fmt.Println("Output:", string(output))
			fmt.Println("Error:", execution["error"])
			break
		}

		fmt.Printf("⏳ Status: %s - polling again in 2s...\\n", status)
		time.Sleep(2 * time.Second)
	}
}`;
}

/**
 * Get display name for language
 */
export function getLanguageDisplayName(language: Language): string {
  switch (language) {
    case 'curl':
      return 'cURL';
    case 'python':
      return 'Python';
    case 'javascript-axios':
      return 'JavaScript (Axios)';
    case 'javascript-fetch':
      return 'JavaScript (Fetch)';
    case 'javascript-ajax':
      return 'JavaScript (jQuery/Ajax)';
    case 'go':
      return 'Go';
    default:
      return language;
  }
}

/**
 * Get syntax highlighting language for Prism/highlight.js
 */
export function getSyntaxLanguage(language: Language): string {
  switch (language) {
    case 'curl':
      return 'bash';
    case 'python':
      return 'python';
    case 'javascript-axios':
    case 'javascript-fetch':
    case 'javascript-ajax':
      return 'javascript';
    case 'go':
      return 'go';
    default:
      return 'text';
  }
}
