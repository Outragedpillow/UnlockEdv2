import{W as c,r as w,j as s,a as u}from"./app-6ded1f35.js";import{G as f}from"./GuestLayout-bcedc2b3.js";import{I as t,T as m,a as n}from"./TextInput-699a9a09.js";import{P as x}from"./PrimaryButton-14133a1e.js";function g(){const{data:e,setData:r,post:d,processing:i,errors:o,reset:p}=c({password:"",password_confirmation:""});w.useEffect(()=>()=>{p("password","password_confirmation")},[]);const l=a=>{a.preventDefault(),d(route("password.update"))};return s.jsxs(f,{children:[s.jsx(u,{title:"Reset Password"}),s.jsxs("form",{onSubmit:l,children:[s.jsxs("div",{className:"mt-4",children:[s.jsx(t,{htmlFor:"password",value:"New Password"}),s.jsx(m,{id:"password",type:"password",name:"password",value:e.password,className:"mt-1 block w-full",autoComplete:"new-password",isFocused:!0,onChange:a=>r("password",a.target.value)}),s.jsx(n,{message:o.password,className:"mt-2"})]}),s.jsxs("div",{className:"mt-4",children:[s.jsx(t,{htmlFor:"password_confirmation",value:"Confirm Password"}),s.jsx(m,{type:"password",name:"password_confirmation",value:e.password_confirmation,className:"mt-1 block w-full",autoComplete:"new-password",onChange:a=>r("password_confirmation",a.target.value)}),s.jsx(n,{message:o.password_confirmation,className:"mt-2"})]}),s.jsx("div",{className:"flex items-center justify-end mt-4",children:s.jsx(x,{className:"ms-4",disabled:i,children:"Reset Password"})})]})]})}export{g as default};