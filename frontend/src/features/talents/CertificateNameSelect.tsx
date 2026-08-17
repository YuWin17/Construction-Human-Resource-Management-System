import { useQuery } from '@tanstack/react-query'
import { AutoComplete } from 'antd'
import { useMemo, useState } from 'react'
import { listCertificateCatalogs } from '../../api/talents'

type Props = {
  value?: string
  onChange?: (value: string) => void
  disabled?: boolean
}

export function CertificateNameSelect({ value, onChange, disabled }: Props) {
  const catalogs = useQuery({ queryKey: ['certificate-catalogs'], queryFn: () => listCertificateCatalogs() })
  const [searchText, setSearchText] = useState('')
  const options = useMemo(() => {
    const keyword = searchText.trim().toLocaleLowerCase()
    return (catalogs.data ?? [])
      .filter((catalog) => !keyword || catalog.name.toLocaleLowerCase().includes(keyword))
      .map((catalog) => ({ value: catalog.name, label: catalog.name }))
  }, [catalogs.data, searchText])

  return <AutoComplete
    allowClear
    value={value}
    disabled={disabled}
    placeholder="输入或搜索证书名称"
    notFoundContent={catalogs.isLoading ? '加载中...' : '可直接输入新证书名称'}
    filterOption={false}
    options={options}
    onSearch={setSearchText}
    onChange={(nextValue) => { setSearchText(nextValue); onChange?.(nextValue) }}
  />
}
