package main

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"math/rand"
	"os"
)

type field struct {
	Type    string
	Name    string
	ObfName string
}

type structType struct {
	Package string
	Name    string
	ObfName string
	Fields  []field
	// Funcs    funcs
}

//type structs []*structType

var (
	logDebug = flag.Bool("D", false, "debug")
)

func init() {
	log.SetFlags(log.Flags() | log.Llongfile)
}

func main() {
	flag.Parse()
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "tester2.test", nil, 0)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("\n\n----------------------------------------\n")
	log.Printf("Original AST:\n\n")
	printer.Fprint(os.Stdout, fset, f)
	log.Printf("\n\n----------------------------------------\n\n")

	foundStructs := findDaStructs(f)
	rewriteDaStructs(f, foundStructs)

	log.Printf("\n\n----------------------------------------\n")
	log.Printf("Modified AST:\n\n")
	printer.Fprint(os.Stdout, fset, f)
	log.Printf("\n\n----------------------------------------\n")
}

func findDaStructs(f *ast.File) []*structType {
	foundStructs := []*structType{}

	ast.Inspect(f, func(n ast.Node) bool {
		g, ok := n.(*ast.GenDecl)

		// is this a func?
		if !ok || g.Tok != token.TYPE {
			return true
		}

		st := &structType{}

		for _, spec := range g.Specs {
			t := spec.(*ast.TypeSpec)

			// is this a struct?
			s, ok := t.Type.(*ast.StructType)
			if !ok {
				log.Println("i am not ok")
				return false
			}

			st.Name = t.Name.Name
			for _, f := range s.Fields.List {
				field := field{Name: f.Names[0].Name, Type: ""}
				switch f.Type.(type) {
				case *ast.Ident:
					field.Type = f.Type.(*ast.Ident).Name
				}
				// anic: interface conversion: ast.Expr is *ast.ArrayType, not *ast.Ident
				//field := field{
				//	Name: f.Names[0].Name,
				//	Type: f.Type.(*ast.Ident).Name,
				//}
				st.Fields = append(st.Fields, field)
			}
		}
		//log.Printf("structName: %s\n", st.Name)
		//for _, v := range st.Fields {
		//	log.Printf("\t%s:%s\n", v.Name, v.Type)
		//}
		foundStructs = append(foundStructs, st)
		return true
	})
	return foundStructs

}

func rewriteDaStructs(f *ast.File, s []*structType) {
	for _, v := range s {
		log.Println("\nlooking for " + v.Name + "...")
		myNewStructName := randoStringo(5)
		ast.Inspect(f, func(n ast.Node) bool {
			//log.Printf("type: %T\n", n)
			g, ok := n.(*ast.GenDecl)

			// is this a func?
			if !ok {
				f, fOk := n.(*ast.FuncDecl)
				if !fOk {
					//log.Println("not ok, but not fOk either")
					return true
				}
				if *logDebug {
					log.Printf("func: %s\n", f.Name.Name)
				}
				for _, funcBodyStmt := range f.Body.List {
					if *logDebug {
						log.Printf("found %T\n", funcBodyStmt)
					}
					switch funcBodyStmt.(type) {
					case *ast.AssignStmt:
						fbsAs, _ := funcBodyStmt.(*ast.AssignStmt)
						//	log.Printf("    %T\n", fbsAs.Tok)
						//	log.Printf("	%s\n", fbsAs.Tok.String())
						//TODO can we always use Lhs or does Rhs need to be checked?
						for _, lh := range fbsAs.Lhs {
							if *logDebug {
								log.Printf("	lh: %T\n", lh)
							}
							switch lh.(type) {
							case *ast.Ident:
								lhI, _ := lh.(*ast.Ident)
								if *logDebug {
									log.Println("	lhI.Name: " + lhI.Name)
								}
							default:
								log.Printf("no case found in lh for type %T, maybe you need one?", lh)
							}
						}
						for _, rh := range fbsAs.Rhs {
							if *logDebug {
								log.Printf("	rh: %T\n", rh)
							}
							switch rh.(type) {
							case *ast.UnaryExpr:
								rhU, ok := rh.(*ast.UnaryExpr)
								if !ok {
									log.Printf("tried to cast to a *ast.UnaryExpr, got a: %T\n", rh)
									return false
								}
								if *logDebug {
									log.Printf("	rhU: %v\n", rhU.X)
								}
								rhUC, ok := rhU.X.(*ast.CompositeLit)
								if !ok {
									log.Printf("tried to cast to a *ast.CompositeLit, got a: %T\n", rhU.X)
									return false
								}

								switch rhUC.Type.(type) {
								case *ast.Ident:
									rhUCType, ok := rhUC.Type.(*ast.Ident)
									if !ok {
										log.Printf("tried to cast to a *ast.Ident, got a: %T\n", rhUC.Type)
										return false
									}
									if rhUCType.Name == v.Name {
										rhUCType.Name = myNewStructName
									}
								case *ast.SelectorExpr:
									log.Println("*ast.SelectorExpr test...")
									rhUCType, ok := rhUC.Type.(*ast.SelectorExpr)
									if !ok {
										log.Printf("tried to cast to a *ast.SelectorExpr, got a: %T\n", rhUC.Type)
										return false
									}
									log.Printf("*ast.SelectorExpr name: %s\n", rhUCType.Sel.Name)
									if rhUCType.Sel.Name == v.Name {
										rhUCType.Sel.Name = myNewStructName
									}
								default:
									log.Printf("no case found in rhUC for type %T, maybe you need one?", rhUC)
								}
							case *ast.CompositeLit:
								rhU, _ := rh.(*ast.CompositeLit)
								rhUCType, _ := rhU.Type.(*ast.Ident)
								if rhUCType.Name == v.Name {
									rhUCType.Name = myNewStructName
								}
								if len(rhU.Elts) > 0 {
									if *logDebug {
										log.Println("	party time")
									}
									for _, e := range rhU.Elts {
										if *logDebug {
											log.Printf("\t%T:%v\n", e, e)
										}
										eKV, _ := e.(*ast.KeyValueExpr)
										ei, _ := eKV.Key.(*ast.Ident)
										for _, ek := range v.Fields {
											if ei.Name == ek.Name {
												ei.Name = ek.ObfName
											}
										}
									}
								}
							case *ast.CallExpr:
							//	rhU, _ := rh.(*ast.CallExpr)
							//	log.Printf("Args: %v\n", rhU.Args)
							default:
								log.Printf("no case found in rh for type %T, maybe you need one?", rh)
							}
						}
					case *ast.DeclStmt:
						log.Println("")
						fbsDs, _ := funcBodyStmt.(*ast.DeclStmt)
						log.Printf("ast.DeclStmt.Decl: %T\n", fbsDs.Decl)
						switch fbsDs.Decl.(type) {
						case *ast.GenDecl:
							fbsDsGd, _ := fbsDs.Decl.(*ast.GenDecl)
							log.Printf("entering the mangler with a: %T\n", fbsDsGd.Tok)
							genDeclMangler(fbsDsGd, v, myNewStructName)
							log.Println("back from mangler")
						}
					}
				}
				//if !ok || g.Tok != token.TYPE {
				//	log.Println("	Skipped!")
				return true
			} // HO LEE SHIT !

			// not a func
			genDeclMangler(g, v, myNewStructName)

			return true
		})
	}
}

func genDeclMangler(g *ast.GenDecl, v *structType, myNewStructName string) {
	// not a func
	switch g.Tok {
	case token.TYPE:
		log.Println("i am a token.Type")
		for _, spec := range g.Specs {
			t, ok := spec.(*ast.TypeSpec)
			if !ok {
				log.Printf("tried to cast to a *ast.TypeSpec, got a: %T\n", spec)
			}
			switch t.Type.(type) {
			// is this a struct?
			case *ast.StructType:
				if t.Name.Name == v.Name {
					y, ok := t.Type.(*ast.StructType)
					if !ok {
						log.Println("i am not ok")
						//return false
						return
					}
					t.Name.Name = myNewStructName
					v.ObfName = myNewStructName
					for _, f := range y.Fields.List {
						origName := f.Names[0].Name
						obfName := randoStringo(5)

						f.Names[0].Name = obfName
						for ni, nk := range v.Fields {
							if nk.Name == origName {
								v.Fields[ni].ObfName = obfName
								nk.Name = obfName
							}
						}
					}
				}
			case *ast.Ident:
				log.Printf("notastruct: %s\n", t.Name.Name)
			default:
				log.Printf("no case found in specs for type %T, maybe you need one?\n", spec)
			}
		}
	case token.VAR:
		log.Println("i am a token.VAR")
		log.Printf("length of g.Specs: %d\n", len(g.Specs))
		for _, spec := range g.Specs {
			t, ok := spec.(*ast.ValueSpec)
			if !ok {
				log.Printf("tried to cast to a *ast.ValueSpec, got a: %T\n", t)
			}
			//log.Printf("%v:%v\n", len(t.Names), len(t.Values))
			//log.Printf("%s : %T\n", t.Names[0].Name, t.Values[0])
			log.Printf("length of t.Values: %d\n", len(t.Values))
			log.Printf("t.Type: %v\n", t.Type)
			if len(t.Values) > 0 {
				switch t.Values[0].(type) {
				case *ast.CompositeLit:
					blahVar, ok := t.Values[0].(*ast.CompositeLit)
					if !ok {
						log.Printf("tried to cast to a *ast.CompositeLit, got a: %T\n", t.Values[0])
						//return false
						return
					}

					blahType, ok := blahVar.Type.(*ast.Ident)
					if !ok {
						log.Printf("tried to cast to a *ast.Ident, got a: %T\n", t.Values[0])
						//return false
						return
					}

					if blahType.Name == v.Name {
						//			log.Println("var to replace found")
						blahType.Name = myNewStructName
					}
				case *ast.CallExpr:
					log.Println("CallExpr coming in...")
					blahVar, _ := t.Values[0].(*ast.CallExpr)
					log.Printf("Fun: %v\n", blahVar.Fun)
					// log.Printf("Args: %v\n", rhU.Args)
					if len(blahVar.Args) > 0 {
						for _, blahVarArgs := range blahVar.Args {
							log.Printf("blahVarArgs: %v\n", blahVarArgs)
						}
					}

				default:
					log.Printf("no case found in t.Values[0] for type %T, maybe you need one?", t.Values[0])
				}
			} else {
				//YOLO
				// t.Type is THE type ;)
				log.Printf("type of t.Type: %T\n", t.Type)
				switch t.Type.(type) {
				case *ast.Ident:
					yoloIT, _ := t.Type.(*ast.Ident)
					if yoloIT.Name == v.Name {
						yoloIT.Name = myNewStructName
					}
				default:
					log.Println("Miss")
				}

			}
		}
	default:
		log.Printf("no case found in g.Tok for type %T, maybe you need one?", g.Tok)
	}
}

// TODO: update to honor go naming conventions
func randoStringo(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
