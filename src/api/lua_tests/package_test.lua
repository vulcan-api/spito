function main()
	p, err = api.pkg.Get("bash")

	return err == nil and p.Name ~= "" and p.Version ~= "" and p.InstallDate ~= ""
end
