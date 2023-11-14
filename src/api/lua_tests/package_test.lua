function main()
	p = Package{}
	err = p.Get(p, "bash")

	return err == nil and p.Name ~= "" and p.Version ~= "" and p.InstallDate ~= ""
end
