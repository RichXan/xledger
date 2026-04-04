import { useState, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { Upload, AlertCircle, CheckCircle } from 'lucide-react'
import { useMutation } from '@tanstack/react-query'
import { importCSVPreview, importCSVConfirm } from '@/features/management/management-api'
import type { ImportPreviewResponse } from '@/features/management/management-api'

type ImportStep = 'upload' | 'mapping' | 'result'

export function ImportPage() {
  const { t } = useTranslation()
  const [step, setStep] = useState<ImportStep>('upload')
  const [file, setFile] = useState<File | null>(null)
  const [preview, setPreview] = useState<ImportPreviewResponse | null>(null)
  const [mapping, setMapping] = useState<Record<string, string>>({})
  const fileInputRef = useRef<HTMLInputElement>(null)

  const previewMutation = useMutation({
    mutationFn: async (formData: FormData) => {
      const response = await importCSVPreview(formData)
      return response
    },
    onSuccess: (data) => {
      setPreview(data)
      setMapping(data.suggested_mapping || {})
      setStep('mapping')
    },
  })

  const confirmMutation = useMutation({
    mutationFn: async (mapping: Record<string, string>) => {
      const formData = new FormData()
      if (file) formData.append('file', file)
      formData.append('mapping', JSON.stringify(mapping))
      const response = await importCSVConfirm(formData)
      return response
    },
    onSuccess: () => {
      setStep('result')
    },
  })

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0]
    if (selectedFile) {
      setFile(selectedFile)
      const formData = new FormData()
      formData.append('file', selectedFile)
      previewMutation.mutate(formData)
    }
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    const droppedFile = e.dataTransfer.files[0]
    if (droppedFile && droppedFile.name.endsWith('.csv')) {
      setFile(droppedFile)
      const formData = new FormData()
      formData.append('file', droppedFile)
      previewMutation.mutate(formData)
    }
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <h1 className="text-2xl font-bold">{t('import.title')}</h1>

      {step === 'upload' && (
        <div
          className="border-2 border-dashed border-outline/30 rounded-2xl p-12 text-center cursor-pointer hover:border-primary transition-colors"
          onDrop={handleDrop}
          onDragOver={(e) => e.preventDefault()}
          onClick={() => fileInputRef.current?.click()}
        >
          <Upload className="mx-auto mb-4 text-on-surface-variant" size={48} />
          <p className="text-lg font-medium">{t('import.selectFile')}</p>
          <p className="text-sm text-on-surface-variant mt-2">
            {t('import.supportedFormats')}
          </p>
          <input
            ref={fileInputRef}
            type="file"
            accept=".csv"
            className="hidden"
            onChange={handleFileChange}
          />
        </div>
      )}

      {previewMutation.isPending && (
        <div className="text-center py-8">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          <p className="mt-2 text-on-surface-variant">{t('common.loading')}</p>
        </div>
      )}

      {previewMutation.isError && (
        <div className="flex items-center gap-2 p-4 rounded-xl bg-red-50 text-red-700">
          <AlertCircle size={20} />
          <p>{previewMutation.error instanceof Error ? previewMutation.error.message : t('errors.serverError')}</p>
        </div>
      )}

      {step === 'mapping' && preview && (
        <div className="space-y-6">
          {/* Format detection result */}
          <div className="rounded-xl bg-surface-container p-4">
            <p className="text-sm font-medium">
              {t('import.detectedFormat')}: <span className="text-primary">{preview.format}</span>
            </p>
          </div>

          {/* Preview table */}
          <div className="overflow-x-auto rounded-xl border border-outline/15">
            <table className="w-full text-sm">
              <thead className="bg-surface-container">
                <tr>
                  <th className="px-4 py-2 text-left">#</th>
                  {preview.columns.slice(0, 5).map((h) => (
                    <th key={h} className="px-4 py-2 text-left">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {preview.sample_rows.slice(0, 3).map((row, i) => (
                  <tr key={i} className="border-t border-outline/10">
                    <td className="px-4 py-2 text-on-surface-variant">#{i + 1}</td>
                    {row.slice(0, 5).map((cell, j) => (
                      <td key={j} className="px-4 py-2">{cell}</td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Field mapping */}
          <div className="space-y-3">
            <h3 className="font-semibold">{t('import.fieldMapping')}</h3>
            {Object.entries(mapping).map(([field, csvCol]) => (
              <div key={field} className="flex items-center gap-4">
                <label className="w-32 text-sm font-medium">{field}</label>
                <select
                  value={csvCol}
                  onChange={(e) => setMapping({ ...mapping, [field]: e.target.value })}
                  className="flex-1 rounded-xl border border-outline/15 px-4 py-2"
                >
                  <option value="">--</option>
                  {preview.columns.map((h) => (
                    <option key={h} value={h}>{h}</option>
                  ))}
                </select>
              </div>
            ))}
          </div>

          {/* Action buttons */}
          <div className="flex gap-4">
            <button
              onClick={() => confirmMutation.mutate(mapping)}
              disabled={confirmMutation.isPending}
              className="flex-1 rounded-xl bg-primary py-3 text-white font-semibold disabled:opacity-50"
            >
              {confirmMutation.isPending ? t('common.loading') : t('import.confirm')}
            </button>
            <button
              onClick={() => { setStep('upload'); setFile(null); setPreview(null) }}
              className="flex-1 rounded-xl border border-outline/15 py-3"
            >
              {t('common.cancel')}
            </button>
          </div>
        </div>
      )}

      {step === 'result' && confirmMutation.data && (
        <div className="space-y-6 text-center py-8">
          <CheckCircle className="mx-auto text-emerald-500" size={64} />
          <h2 className="text-xl font-bold">{t('import.success')}</h2>
          <p className="text-on-surface-variant">
            {confirmMutation.data.success_count} {t('import.successRows')}
            {confirmMutation.data.skip_count > 0 && ` • ${confirmMutation.data.skip_count} ${t('import.skipped')}`}
            {confirmMutation.data.fail_count > 0 && ` • ${confirmMutation.data.fail_count} ${t('import.failed')}`}
          </p>
          <button
            onClick={() => { setStep('upload'); setFile(null); setPreview(null) }}
            className="rounded-xl bg-primary px-8 py-3 text-white font-semibold"
          >
            {t('import.importMore')}
          </button>
        </div>
      )}
    </div>
  )
}
